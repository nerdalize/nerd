package client

import (
	"path"

	"github.com/dghubble/sling"
	"github.com/nerdalize/nerd/nerd/client"
	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/nerdalize/nerd/nerd/payload"
	"github.com/pkg/errors"
)

const (
	//AuthHeader is the name of the HTTP Authorization header.
	AuthHeader = "Authorization"

	projectsPrefix = "projects"
)

//NerdAPIClient is a client for the Nerdalize API.
type NerdAPIClient struct {
	client.NerdAPIConfig
}

//NewNerdAPI creates a new NerdAPIClient from a config object. When no URL is set it tries to get the URL from the audience field in the JWT.
func NewNerdAPI(conf client.NerdAPIConfig) (*NerdAPIClient, error) {
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
func (nerdapi *NerdAPIClient) url(p string) string {
	return nerdapi.URL + "/" + path.Join(projectsPrefix, nerdapi.ProjectID, p)
}

//doRequest makes the actual request. First it fetches the credentials (nerd token) and then it creates the request to the API server.
//doRequest checks if the server responded with a payload error and hands this error back to the user.
func (nerdapi *NerdAPIClient) doRequest(s *sling.Sling, result interface{}) error {
	value, err := nerdapi.Credentials.Get()
	if err != nil {
		return errors.Wrap(err, "failed to get credentials")
	}
	e := &payload.Error{}
	req, err := s.Request()
	if err != nil {
		return errors.Wrap(err, "could not create request")
	}
	req.Header.Add(AuthHeader, "Bearer "+value.NerdToken)
	// logRequest(req) @TODO inject a logger into the client
	res, err := s.Receive(result, e)
	if err != nil {
		return errors.Wrapf(err, "unexpected behaviour when making request to %v (%v), with headers (%v)", req.URL, req.Method, req.Header)
	}
	// logResponse(res) @TODO inject a logger into client
	if e.Message != "" {
		return &HTTPError{
			StatusCode: res.StatusCode,
			Err:        e,
		}
	}
	return nil
}
