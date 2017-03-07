package client

import (
	"path"

	"github.com/dghubble/sling"
	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/nerdalize/nerd/nerd/payload"
	"github.com/pkg/errors"
)

const (
	AuthHeader = "Authorization"

	projectsPrefix = "projects/6de308f4-face-11e6-bc64-92361f002671"

	tasksEndpoint    = "tasks"
	sessionsEndpoint = "tokens"
	workersEndpoint  = "workers"
)

//NerdAPIClient is a client for the Nerdalize API.
type NerdAPIClient struct {
	URL         string
	Credentials *credentials.NerdAPI
}

func NewNerdAPI(cred *credentials.NerdAPI) (*NerdAPIClient, error) {
	value, err := cred.Get()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get credentials")
	}
	claims, err := credentials.DecodeToken(value.NerdToken)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode token '%v'", value.NerdToken)
	}

	if claims.Audience == "" {
		return nil, errors.Errorf("nerd token '%v' does not contain audience field", claims.Audience)
	}
	return NewNerdAPIWithEndpoint(cred, claims.Audience), nil
}

func NewNerdAPIWithEndpoint(cred *credentials.NerdAPI, url string) *NerdAPIClient {
	return &NerdAPIClient{
		Credentials: cred,
		URL:         url,
	}
}

//url returns the full endpoint url appended with a given path.
func (nerdapi *NerdAPIClient) url(p string) (string, error) {
	value, err := nerdapi.Credentials.Get()
	if err != nil {
		return "", errors.Wrap(err, "failed to get credentials")
	}
	claims, err := credentials.DecodeToken(value.NerdToken)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode token")
	}
	return nerdapi.URL + "/" + path.Join(projectsPrefix, claims.ProjectID, p), nil
}

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
	url, err := nerdapi.url(path.Join(workersEndpoint))
	if err != nil {
		return nil, err
	}

	s := sling.New().Post(url)
	err = nerdapi.doRequest(s, worker)
	return
}

//DeleteWorker removes a worker
func (nerdapi *NerdAPIClient) DeleteWorker(workerID string) (err error) {
	url, err := nerdapi.url(path.Join(workersEndpoint, workerID))
	if err != nil {
		return err
	}

	s := sling.New().Delete(url)
	err = nerdapi.doRequest(s, nil)
	return
}

//CreateSession creates a new user session.
func (nerdapi *NerdAPIClient) CreateSession() (sess *payload.SessionCreateOutput, err error) {
	sess = &payload.SessionCreateOutput{}
	url, err := nerdapi.url(path.Join(sessionsEndpoint))
	if err != nil {
		return nil, err
	}
	s := sling.New().Post(url)
	err = nerdapi.doRequest(s, sess)
	return
}

//CreateTask creates a new executable task.
func (nerdapi *NerdAPIClient) CreateTask(image string, dataset string, args []string) (output *payload.TaskCreateOutput, err error) {
	output = &payload.TaskCreateOutput{}
	// create payload
	p := &payload.TaskCreateInput{
		Image: image,
	}

	// post request
	url, err := nerdapi.url(tasksEndpoint)
	if err != nil {
		return nil, err
	}

	s := sling.New().
		Post(url).
		BodyJSON(p)

	err = nerdapi.doRequest(s, output)
	return
}

//PatchTaskStatus updates the status of a task.
func (nerdapi *NerdAPIClient) PatchTaskStatus(id string, ts *payload.TaskCreateInput) error {
	ts = &payload.TaskCreateInput{}
	url, err := nerdapi.url(path.Join(tasksEndpoint, id))
	if err != nil {
		return err
	}
	s := sling.New().
		Patch(url).
		BodyJSON(ts)

	return nerdapi.doRequest(s, nil)
}

//ListTasks lists all tasks.
func (nerdapi *NerdAPIClient) ListTasks() (tl *payload.TaskListOutput, err error) {
	tl = &payload.TaskListOutput{}
	url, err := nerdapi.url(tasksEndpoint)
	if err != nil {
		return nil, err
	}
	s := sling.New().Get(url)
	err = nerdapi.doRequest(s, tl)
	return
}
