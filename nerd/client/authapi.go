package client

import (
	"github.com/dghubble/sling"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/nerd/payload"
	"github.com/pkg/errors"
)

const (
	//TokenEndpoint is the endpoint from where to fetch the JWT.
	TokenEndpoint = "token/?service=nce.nerdalize.com" // TODO: Make configurable
)

//AuthAPIClient is the client for the nerdalize authentication server.
type AuthAPIClient struct {
	Config conf.AuthConfig
}

//NewAuthAPI creates a new AuthAPIClient.
func NewAuthAPI(config conf.AuthConfig) *AuthAPIClient {
	return &AuthAPIClient{
		Config: config,
	}
}

// GetOAuthToken requests a oauth token from the auth endpoint.
func (auth *AuthAPIClient) GetOAuthToken(code string) (*payload.OAuthTokens, error) {
	type formbody struct {
		Code        string `url:"code"`
		ClientID    string `url:"client_id"`
		RedirectURI string `url:"redirect_uri"`
		GrantType   string `url:"grant_type"`
	}

	tokenParams := &formbody{
		Code:        code,
		ClientID:    auth.Config.ClientID,
		GrantType:   "authorization_code",
		RedirectURI: "http://" + auth.Config.OAuthLocalserver + "/oauth/callback",
	}
	s := sling.New().Post(auth.Config.APIEndpoint + "/o/token/").BodyForm(tokenParams)
	s = s.Set("Content-Type", "application/x-www-form-urlencoded")
	s = s.Set("Accept", "application/json")

	req, err := s.Request()
	if err != nil {
		return &payload.OAuthTokens{}, errors.Wrapf(err, "failed to create request (%v)", auth.Config.APIEndpoint)
	}

	tokens := &payload.OAuthTokens{}
	e := &payload.OAuthError{}
	// e := make(map[string]string)
	_, err = s.Do(req, tokens, e)

	if err != nil {
		return &payload.OAuthTokens{}, errors.Wrapf(err, "failed to do request (%v)", auth.Config.APIEndpoint)
	}
	if e.OAuthError != "" {
		return &payload.OAuthTokens{}, e
	}

	return tokens, nil

}
