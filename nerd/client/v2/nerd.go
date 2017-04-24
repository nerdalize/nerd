package v2client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/nerdalize/nerd/nerd/payload"
)

const (
	//AuthHeader is the name of the HTTP Authorization header.
	AuthHeader = "Authorization"

	projectsPrefix = "projects"

	tasksEndpoint    = "tasks"
	sessionsEndpoint = "tokens"
	datasetEndpoint  = "datasets"
	workersEndpoint  = "workers"
)

//Nerd is a client for the Nerdalize API.
type Nerd struct {
	NerdConfig
	cred *Credentials
}

//NerdConfig provides config details to create a Nerd client.
type NerdConfig struct {
	Client              Doer
	CredentialsProvider CredentialsProvider
	Base                *url.URL
	Logger              Logger
	// AWSClient           AWSClient
}

// Doer executes http requests.  It is implemented by *http.Client.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Credentials contains the JWT required to authenticate to the nerd API.
type Credentials struct {
	JWT string
}

//CredentialsProvider provides the client with credentials. An implementation of this interface
//is capable of providing a Credentials object to the client. When IsExpired return false
//an in-memory Credentials object will be used to prevent from calling Retrieve for each API call.
type CredentialsProvider interface {
	IsExpired() bool
	Retrieve() (*Credentials, error)
}

//AWSClient sends requests to AWS.
// type AWSClient interface {
// 	PutObject(key string, body io.ReadCloser) (err error)
// }

//NewNerdClient creates a new Nerd client from a config object. The http.DefaultClient
//will be used as default Doer.
func NewNerdClient(conf NerdConfig) *Nerd {
	if conf.Client == nil {
		conf.Client = http.DefaultClient
	}
	if conf.Base.Path != "" && conf.Base.Path[len(conf.Base.Path)-1] != '/' {
		conf.Base.Path = conf.Base.Path + "/"
	}
	cl := &Nerd{
		NerdConfig: conf,
		cred:       nil,
	}
	return cl
}

func (c *Nerd) getCredentials() (*Credentials, error) {
	if c.CredentialsProvider == nil {
		return nil, fmt.Errorf("Please provide a credentials provider")
	}
	if c.cred == nil || c.CredentialsProvider.IsExpired() {
		cred, err := c.CredentialsProvider.Retrieve()
		if err != nil {
			return nil, err
		}
		c.cred = cred
	}
	return c.cred, nil
}

func (c *Nerd) doRequest(method, urlPath string, input, output interface{}) (err error) {
	cred, err := c.getCredentials()
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	if input != nil {
		enc := json.NewEncoder(buf)
		err = enc.Encode(input)
		if err != nil {
			return fmt.Errorf("failed to encode the request body: %+v", err)
		}
	}

	path, err := url.Parse(urlPath)
	if err != nil {
		return fmt.Errorf("invalid url path provided: %+v", err)
	}

	resolved := c.Base.ResolveReference(path)
	req, err := http.NewRequest(method, resolved.String(), buf)
	logRequest(req, c.Logger)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %+v", err)
	}

	req.Header.Set("Authorization", "Bearer "+cred.JWT)
	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform HTTP request: %+v", err)
	}
	logResponse(resp, c.Logger)

	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		errv := &payload.Error{}
		err = dec.Decode(errv)
		if err != nil {
			return fmt.Errorf("failed to decode unexpected HTTP response (%s): %+v", resp.Status, err)
		}

		return &HTTPError{
			StatusCode: resp.StatusCode,
			Err:        errv,
		}
		// return fmt.Errorf("unexpected response HTTP response: %s, error: %#v", resp.Status, errv)
	}

	if output != nil {
		err = dec.Decode(output)
		if err != nil {
			return fmt.Errorf("failed to decode successfull HTTP response (%s): %+v", resp.Status, err)
		}
	}

	return nil
}

func createPath(projectID string, elem ...string) string {
	return "projects/" + projectID + "/" + path.Join(elem...)
}

//CreateWorker creates registers this client as workable capacity
func (c *Nerd) CreateWorker(projectID string) (output *payload.WorkerCreateOutput, err error) {
	output = &payload.WorkerCreateOutput{}
	return output, c.doRequest(http.MethodPost, createPath(projectID, workersEndpoint), nil, output)
}

// DeleteWorker removes a worker
func (c *Nerd) DeleteWorker(projectID, workerID string) (err error) {
	return c.doRequest(http.MethodDelete, createPath(projectID, workersEndpoint, workerID), nil, nil)
}

//CreateSession creates a new user session.
func (c *Nerd) CreateSession(projectID string) (output *payload.SessionCreateOutput, err error) {
	// logrus.Debug("Creating session")
	output = &payload.SessionCreateOutput{}
	return output, c.doRequest(http.MethodPost, createPath(projectID, sessionsEndpoint), nil, output)
}

//CreateTask creates a new executable task.
func (c *Nerd) CreateTask(projectID, image, dataset string, env map[string]string) (output *payload.TaskCreateOutput, err error) {
	// logrus.WithFields(logrus.Fields{
	// 	"image":   image,
	// 	"dataset": dataset,
	// }).Debug("Creating task")
	output = &payload.TaskCreateOutput{}
	// create payload
	input := &payload.TaskCreateInput{
		Image:       image,
		InputID:     dataset,
		Environment: env,
	}
	return output, c.doRequest(http.MethodPost, createPath(projectID, tasksEndpoint), input, output)
}

//PatchTaskStatus updates the status of a task.
func (c *Nerd) PatchTaskStatus(projectID, id string, input *payload.TaskCreateInput) error {
	// logrus.WithFields(logrus.Fields{
	// 	"id": id,
	// }).Debug("Patching task")
	return c.doRequest(http.MethodPatch, createPath(projectID, tasksEndpoint, id), input, nil)
}

//ListTasks lists all tasks.
func (c *Nerd) ListTasks(projectID string) (output *payload.TaskListOutput, err error) {
	// logrus.Debug("Listing tasks")
	output = &payload.TaskListOutput{}
	return output, c.doRequest(http.MethodGet, createPath(projectID, tasksEndpoint), nil, output)
}

//CreateDataset creates a new dataset.
func (c *Nerd) CreateDataset(projectID string) (output *payload.DatasetCreateOutput, err error) {
	output = &payload.DatasetCreateOutput{}
	return output, c.doRequest(http.MethodPost, createPath(projectID, datasetEndpoint), nil, output)
}

//GetDataset gets a dataset by ID.
func (c *Nerd) GetDataset(projectID, id string) (output *payload.DatasetDescribeOutput, err error) {
	output = &payload.DatasetDescribeOutput{}
	return output, c.doRequest(http.MethodGet, createPath(projectID, datasetEndpoint, id), nil, output)
}
