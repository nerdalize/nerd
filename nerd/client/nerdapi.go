package client

import (
	"fmt"
	"net/url"
	"path"

	"github.com/dghubble/sling"
	"github.com/nerdalize/nerd/nerd/payload"
	"github.com/pkg/errors"
)

const (
	defaultScheme   = "https"
	defaultHost     = "platform.nerdalize.net"
	defaultBasePath = ""
	defaultVersion  = "v1"

	tasksEndpoint    = "tasks"
	sessionsEndpoint = "sessions"
)

//NerdAPIClient is a client for the Nerdalize API.
type NerdAPIClient struct {
	NerdAPIConfig
}

//NerdAPIConfig contains the information needed to create a NerdAPIClient.
type NerdAPIConfig struct {
	Scheme   string
	Host     string
	BasePath string
	Version  string
}

//NewNerdAPI returns a new NerdAPIClient according to a given configuration.
func NewNerdAPI(config NerdAPIConfig) *NerdAPIClient {
	if config.Scheme == "" {
		config.Scheme = defaultScheme
	}
	if config.Host == "" {
		config.Host = defaultHost
	}
	if config.BasePath == "" {
		config.BasePath = defaultBasePath
	}
	if config.Version == "" {
		config.Version = defaultVersion
	}
	return &NerdAPIClient{
		NerdAPIConfig: config,
	}
}

//NewNerdAPIFromURL returns a new NerdAPIClient given a full endpoint URL.
func NewNerdAPIFromURL(fullURL string, version string) (*NerdAPIClient, error) {
	u, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse url '%v': %v", fullURL, err)
	}
	return &NerdAPIClient{
		NerdAPIConfig: NerdAPIConfig{
			Scheme:   u.Scheme,
			Host:     u.Host,
			BasePath: u.Path,
			Version:  version,
		},
	}, nil
}

//url returns the full endpoint url appended with a given path.
func (nerdapi *NerdAPIClient) url(p string) string {
	url := &url.URL{
		Scheme: nerdapi.Scheme,
		Host:   nerdapi.Host,
		Path:   path.Join(nerdapi.BasePath, p),
		//TODO: include version
		// Path:   path.Join(nerdapi.BasePath, nerdapi.Version, p),
	}
	return url.String()
}

func (nerdapi *NerdAPIClient) doRequest(s *sling.Sling, successV interface{}) *APIError {
	e := &payload.Error{}
	req, err := s.Request()
	if err != nil {
		//TODO: should error message include more details like URL, HTTP method and payload (sling is not very verbose in giving detailed error information)?
		return &APIError{
			Response: nil,
			Request:  nil,
			Err:      errors.Wrap(err, "could not create request"),
		}
	}
	resp, err := s.Receive(successV, e)
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

//CreateSession creates a new user session.
func (nerdapi *NerdAPIClient) CreateSession(token string) (sess *payload.Session, err *APIError) {
	url := nerdapi.url(path.Join(sessionsEndpoint, token))
	s := sling.New().Post(url)
	err = nerdapi.doRequest(s, sess)
	return
}

//CreateTask creates a new executable task.
func (nerdapi *NerdAPIClient) CreateTask(image string, dataset string, awsAccessKey string, awsSecret string, args []string) *APIError {
	// set env variables
	args = append(args, "-e=DATASET="+dataset)
	args = append(args, "-e=AWS_ACCESS_KEY_ID="+awsAccessKey)
	args = append(args, "-e=AWS_SECRET_ACCESS_KEY="+awsSecret)

	// create payload
	p := &payload.Task{
		Image:   image,
		Dataset: dataset,
		Args:    args,
	}

	// post request
	url := nerdapi.url(tasksEndpoint)
	s := sling.New().
		Post(url).
		BodyJSON(p)

	return nerdapi.doRequest(s, nil)
}

//PatchTaskStatus updates the status of a task.
func (nerdapi *NerdAPIClient) PatchTaskStatus(id string, ts *payload.TaskStatus) *APIError {
	url := nerdapi.url(path.Join(tasksEndpoint, id))
	s := sling.New().
		Patch(url).
		BodyJSON(ts)

	return nerdapi.doRequest(s, nil)
}

//ListTaskLogs lists the logs of a task.
func (nerdapi *NerdAPIClient) ListTaskLogs(id string) ([]string, *APIError) {
	url := nerdapi.url(path.Join(tasksEndpoint, id))
	t := &payload.Task{}
	s := sling.New().Get(url)
	err := nerdapi.doRequest(s, t)
	return t.LogLines, err
}

//ListTasks lists all tasks.
func (nerdapi *NerdAPIClient) ListTasks() (t []payload.Task, err *APIError) {
	url := nerdapi.url(tasksEndpoint)
	s := sling.New().Get(url)
	err = nerdapi.doRequest(s, &t)
	return
}
