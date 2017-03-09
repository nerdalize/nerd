package client

import (
	"github.com/dghubble/sling"
	"github.com/pkg/errors"
)

const (
	//TokenEndpoint is the endpoint from where to fetch the JWT.
	TokenEndpoint = "token/?service=nce.nerdalize.com"
)

//AuthAPIClient is the client for the nerdalize authentication server.
type AuthAPIClient struct {
	URL string
}

//NewAuthAPI creates a new AuthAPIClient.
func NewAuthAPI(url string) *AuthAPIClient {
	return &AuthAPIClient{
		URL: url,
	}
}

//GetToken gets a JWT for a given user.
func (auth *AuthAPIClient) GetToken(user, pass string) (string, error) {
	type body struct {
		Token string `json:"token"`
	}
	b := &body{}
	s := sling.New().Get(auth.URL + "/" + TokenEndpoint)
	req, err := s.Request()
	if err != nil {
		return "", errors.Wrapf(err, "failed to create request (%v)", auth.URL)
	}
	req.SetBasicAuth(user, pass)
	_, err = s.Do(req, b, nil)
	if err != nil {
		return "", errors.Wrapf(err, "failed to do request (%v)", auth.URL)
	}
	return b.Token, nil
}
