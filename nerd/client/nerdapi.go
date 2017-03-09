package client

import (
	"path"

	"github.com/dghubble/sling"
	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/nerdalize/nerd/nerd/payload"
	"github.com/pkg/errors"
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

//NerdAPIClient is a client for the Nerdalize API.
type NerdAPIClient struct {
	NerdAPIConfig
}

//NerdAPIConfig provides config details to create a NerdAPIClient.
type NerdAPIConfig struct {
	Credentials *credentials.NerdAPI
	URL         string
	ProjectID   string
}

//NewNerdAPI creates a new NerdAPIClient from a config object. When no URL is set
//it tries to get the URL from the audience field in the JWT.
func NewNerdAPI(conf NerdAPIConfig) (*NerdAPIClient, error) {
	cl := &NerdAPIClient{
		conf,
	}
	if cl.URL == "" {
		aud, err := getAudience(conf.Credentials)
		if err != nil {
			// TODO: make it a user facing err
			return nil, errors.Wrap(err, "no valid URL was provided")
		}
		cl.URL = aud
	}
	return cl, nil
}

//getAudience gets the audience from the JWT.
func getAudience(cred *credentials.NerdAPI) (string, error) {
	if cred == nil {
		return "", errors.New("credentials object was nil")
	}
	claims, err := cred.GetClaims()
	if err != nil {
		return "", errors.Wrap(err, "failed to retreive nerd claims")
	}
	if claims.Audience == "" {
		return "", errors.Errorf("nerd token '%v' does not contain audience field", claims.Audience)
	}
	return claims.Audience, nil
}

//url returns the full endpoint url appended with a given path.
func (nerdapi *NerdAPIClient) url(p string) string {
	return nerdapi.URL + "/" + path.Join(projectsPrefix, nerdapi.ProjectID, p)
}

//doRequest makes the actual request. First it fetches the credentials (nerd token) and then it creates the request to the API server.
//doRequest checks if the server responded with a payload error and hands this error back to the user.
func (nerdapi *NerdAPIClient) doRequest(s *sling.Sling, result interface{}) error {
	value, err := nerdapi.Credentials.Get()
	if err != nil {
		// TODO: Is return err ok?
		return &APIError{
			Response: nil,
			Request:  nil,
			Err:      errors.Wrap(err, "failed to get credentials"),
		}
	}
	e := &payload.Error{}
	req, err := s.Request()
	if err != nil {
		return &APIError{
			Response: nil,
			Request:  nil,
			Err:      errors.Wrap(err, "could not create request"),
		}
	}
	req.Header.Add(AuthHeader, "Bearer "+value.NerdToken)
	resp, err := s.Receive(result, e)
	if err != nil {
		return &APIError{
			Response: nil,
			Request:  req,
			Err:      errors.Wrapf(err, "unexpected behaviour when making request to %v (%v), with headers (%v)", req.URL, req.Method, req.Header),
		}
	}
	if e.Message != "" {
		return &APIError{
			Response: resp,
			Request:  req,
			Err:      e,
		}
	}
	return nil
}

//CreateWorker creates registers this client as workable capacity
func (nerdapi *NerdAPIClient) CreateWorker() (worker *payload.WorkerCreateOutput, err error) {
	worker = &payload.WorkerCreateOutput{}
	url := nerdapi.url(path.Join(workersEndpoint))
	s := sling.New().Post(url)
	err = nerdapi.doRequest(s, worker)
	return
}

//DeleteWorker removes a worker
func (nerdapi *NerdAPIClient) DeleteWorker(workerID string) (err error) {
	url := nerdapi.url(path.Join(workersEndpoint, workerID))
	s := sling.New().Delete(url)
	err = nerdapi.doRequest(s, nil)
	return
}

//CreateSession creates a new user session.
func (nerdapi *NerdAPIClient) CreateSession() (sess *payload.SessionCreateOutput, err error) {
	sess = &payload.SessionCreateOutput{}
	url := nerdapi.url(path.Join(sessionsEndpoint))
	s := sling.New().Post(url)
	err = nerdapi.doRequest(s, sess)
	return
}

//CreateTask creates a new executable task.
func (nerdapi *NerdAPIClient) CreateTask(image string, dataset string, env map[string]string) (output *payload.TaskCreateOutput, err error) {
	output = &payload.TaskCreateOutput{}
	// create payload
	p := &payload.TaskCreateInput{
		Image:       image,
		InputID:     dataset,
		Environment: env,
	}

	// post request
	url := nerdapi.url(tasksEndpoint)
	s := sling.New().
		Post(url).
		BodyJSON(p)

	err = nerdapi.doRequest(s, output)
	return
}

//PatchTaskStatus updates the status of a task.
func (nerdapi *NerdAPIClient) PatchTaskStatus(id string, ts *payload.TaskCreateInput) error {
	ts = &payload.TaskCreateInput{}
	url := nerdapi.url(path.Join(tasksEndpoint, id))
	s := sling.New().
		Patch(url).
		BodyJSON(ts)

	return nerdapi.doRequest(s, nil)
}

//ListTasks lists all tasks.
func (nerdapi *NerdAPIClient) ListTasks() (tl *payload.TaskListOutput, err error) {
	tl = &payload.TaskListOutput{}
	url := nerdapi.url(tasksEndpoint)
	s := sling.New().Get(url)
	err = nerdapi.doRequest(s, tl)
	return
}

//CreateDataset creates a new dataset.
func (nerdapi *NerdAPIClient) CreateDataset() (d *payload.DatasetCreateOutput, err error) {
	d = &payload.DatasetCreateOutput{}
	url := nerdapi.url(datasetEndpoint)
	s := sling.New().Post(url)
	err = nerdapi.doRequest(s, d)
	return
}

//GetDataset gets a dataset by ID.
func (nerdapi *NerdAPIClient) GetDataset(id string) (d *payload.DatasetDescribeOutput, err error) {
	d = &payload.DatasetDescribeOutput{}
	url := nerdapi.url(path.Join(datasetEndpoint, id))
	s := sling.New().Get(url)
	err = nerdapi.doRequest(s, d)
	return
}
