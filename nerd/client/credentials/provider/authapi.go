package provider

import (
	"crypto/ecdsa"
	"time"

	"github.com/nerdalize/nerd/nerd/client"
	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

//AuthAPI provides nerdalize credentials by making a request to the nerdalize auth server.
//The UserPassProvider is used to retreive the username and password required to authenticate with the auth server.
type AuthAPI struct {
	*Basis

	Client *client.AuthAPIClient
	//UserPassProvider is a function that provides the AuthAPI provider with a username and password. This could for example be a function
	//that reads from stdin.
	UserPassProvider func() (string, string, error)
}

//NewAuthAPI creates a new AuthAPI provider.
func NewAuthAPI(userPassProvider func() (string, string, error), c *client.AuthAPIClient) *AuthAPI {
	return &AuthAPI{
		Basis: &Basis{
			ExpireWindow: DefaultExpireWindow,
		},
		UserPassProvider: userPassProvider,
		Client:           c,
	}
}

//Retrieve retrieves the token from the authentication server.
func (p *AuthAPI) Retrieve(pub *ecdsa.PublicKey) (*credentials.NerdAPIValue, error) {
	user, pass, err := p.UserPassProvider()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get username or password")
	}
	token, err := p.Client.GetToken(user, pass)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get nerd token for username and password")
	}
	claims, err := credentials.DecodeTokenWithKey(token, pub)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retreive claims from nerd token '%v'", token)
	}
	err = conf.WriteNerdToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to write nerd token to config")
	}
	p.AlwaysValid = claims.ExpiresAt == 0 // if unset
	p.SetExpiration(time.Unix(claims.ExpiresAt, 0))
	return &credentials.NerdAPIValue{
		NerdToken: token,
	}, nil
}
