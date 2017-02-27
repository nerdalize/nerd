package client

import (
	"github.com/dghubble/sling"
	"github.com/pkg/errors"
)

type AuthAPIClient struct {
	URL string
}

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
	s := sling.New().Get(auth.URL)
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
