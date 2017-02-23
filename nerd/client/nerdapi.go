package client

import (
	"path"

	"github.com/dghubble/sling"
	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/nerdalize/nerd/nerd/payload"
	"github.com/pkg/errors"
)

const (
	// TODO: remove these
	defaultScheme   = "https"
	defaultHost     = "platform.nerdalize.net"
	defaultBasePath = ""
	defaultVersion  = "v1"

	AuthHeader = "Authorization"

	tasksEndpoint    = "tasks"
	sessionsEndpoint = "sessions"
)

//NerdAPIClient is a client for the Nerdalize API.
type NerdAPIClient struct {
	URL         string
	Credentials *credentials.NerdAPI
}

//NerdAPIConfig contains the information needed to create a NerdAPIClient.
type NerdAPIConfig struct {
	// Scheme   string
	// Host     string
	// BasePath string
	// Version  string
}

//NewNerdAPI returns a new NerdAPIClient according to a given configuration.
// func NewNerdAPI(config NerdAPIConfig) *NerdAPIClient {
// 	if config.Scheme == "" {
// 		config.Scheme = defaultScheme
// 	}
// 	if config.Host == "" {
// 		config.Host = defaultHost
// 	}
// 	if config.BasePath == "" {
// 		config.BasePath = defaultBasePath
// 	}
// 	if config.Version == "" {
// 		config.Version = defaultVersion
// 	}
// 	return &NerdAPIClient{
// 		NerdAPIConfig: config,
// 	}
// }

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

//NewNerdAPIFromURL returns a new NerdAPIClient given a full endpoint URL.
// func NewNerdAPIFromURL(fullURL string, version string) (*NerdAPIClient, error) {
// 	u, err := url.Parse(fullURL)
// 	if err != nil {
// 		return nil, errors.Wrapf(err, "could not parse url '%v': %v", fullURL)
// 	}
// 	return &NerdAPIClient{
// 		NerdAPIConfig: NerdAPIConfig{
// 			Scheme:   u.Scheme,
// 			Host:     u.Host,
// 			BasePath: u.Path,
// 			Version:  version,
// 		},
// 	}, nil
// }

//url returns the full endpoint url appended with a given path.
func (nerdapi *NerdAPIClient) url(p string) string {
	return nerdapi.URL + "/" + p
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
	req.Header.Add(AuthHeader, "Bearer "+value.NerdToken)
	if err != nil {
		//TODO: should error message include more details like URL, HTTP method and payload (sling is not very verbose in giving detailed error information)?
		return &APIError{
			Response: nil,
			Request:  nil,
			Err:      errors.Wrap(err, "could not create request"),
		}
	}
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
func (nerdapi *NerdAPIClient) CreateSession(token string) (sess *payload.SessionCreateOutput, err error) {
	sess = &payload.SessionCreateOutput{}
	url := nerdapi.url(path.Join(sessionsEndpoint, token))
	s := sling.New().Post(url)
	err = nerdapi.doRequest(s, sess)
	return
}

//CreateTask creates a new executable task.
func (nerdapi *NerdAPIClient) CreateTask(image string, dataset string, args []string) (output *payload.TaskCreateOutput, err error) {
	// set env variables
	// args = append(args, "-e=DATASET="+dataset)
	// args = append(args, "-e=AWS_ACCESS_KEY_ID="+awsAccessKey)
	// args = append(args, "-e=AWS_SECRET_ACCESS_KEY="+awsSecret)
	// _ = args //@TODO fetch these via the API itself
	output = &payload.TaskCreateOutput{}
	// create payload
	p := &payload.TaskCreateInput{
		Image: image,
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
