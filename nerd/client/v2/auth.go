package v2client

import (
	"github.com/dghubble/sling"
	"github.com/nerdalize/nerd/nerd/payload"
	"github.com/pkg/errors"
)

const (
	//TokenEndpoint is the endpoint from where to fetch the JWT.
	TokenEndpoint = "token/?service=nce.nerdalize.com"
)

//Auth is the client for the nerdalize authentication server.
type Auth struct {
	URL string
}

//NewAuthAPI creates a new Auth.
func NewAuthAPI(url string) *Auth {
	return &Auth{
		URL: url,
	}
}

//GetToken gets a JWT for a given user.
func (auth *Auth) GetToken(user, pass string) (string, error) {
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
	e := &payload.AuthError{}
	_, err = s.Do(req, b, e)
	if err != nil {
		return "", errors.Wrapf(err, "failed to do request (%v)", auth.URL)
	}
	if e.Detail != "" {
		return "", e
	}
	return b.Token, nil
}
