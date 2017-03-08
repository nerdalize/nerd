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
	datasetEndpoint  = "datasets"
)

//NerdAPIClient is a client for the Nerdalize API.
type NerdAPIClient struct {
	NerdAPIConfig
}

type NerdAPIConfig struct {
	Credentials *credentials.NerdAPI
	URL         string
}

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
func (nerdapi *NerdAPIClient) url(p string) (string, error) {
	// claims, err := nerdapi.Credentials.GetClaims()
	// if err != nil {
	// 	return "", errors.Wrap(err, "failed to retreive nerd claims")
	// }
	// return nerdapi.URL + "/" + path.Join(projectsPrefix, claims.ProjectID, p), nil

	// TODO: Don't hardcode the ProjectID
	return nerdapi.URL + "/" + path.Join(projectsPrefix, p), nil
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
func (nerdapi *NerdAPIClient) CreateTask(image string, dataset string, env map[string]string) (output *payload.TaskCreateOutput, err error) {
	output = &payload.TaskCreateOutput{}
	// create payload
	p := &payload.TaskCreateInput{
		Image:       image,
		InputID:     dataset,
		Environment: env,
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

//CreateDataset creates a new dataset.
func (nerdapi *NerdAPIClient) CreateDataset() (d *payload.DatasetCreateOutput, err error) {
	d = &payload.DatasetCreateOutput{}
	url, err := nerdapi.url(datasetEndpoint)
	if err != nil {
		return nil, err
	}
	s := sling.New().Post(url)
	err = nerdapi.doRequest(s, d)
	return
}

//GetDataset gets a dataset by ID.
func (nerdapi *NerdAPIClient) GetDataset(id string) (d *payload.DatasetDescribeOutput, err error) {
	d = &payload.DatasetDescribeOutput{}
	url, err := nerdapi.url(path.Join(datasetEndpoint, id))
	if err != nil {
		return nil, err
	}
	s := sling.New().Get(url)
	err = nerdapi.doRequest(s, d)
	return
}
