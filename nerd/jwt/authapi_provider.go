package jwt

import (
	"crypto/ecdsa"

	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	v1payload "github.com/nerdalize/nerd/nerd/client/auth/v1/payload"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

//AuthAPIProvider provides nerdalize credentials by making a request to the nerdalize auth server.
//The UserPassProvider is used to retreive the username and password required to authenticate with the auth server.
type AuthAPIProvider struct {
	*ProviderBasis

	Client *v1auth.Client
	//UserPassProvider is a function that provides the AuthAPIProvider provider with a username and password. This could for example be a function
	//that reads from stdin.
	UserPassProvider func() (string, string, error)
}

//NewAuthAPIProvider creates a new AuthAPIProvider provider.
func NewAuthAPIProvider(pub *ecdsa.PublicKey, userPassProvider func() (string, string, error), c *v1auth.Client) *AuthAPIProvider {
	return &AuthAPIProvider{
		ProviderBasis: &ProviderBasis{
			ExpireWindow: DefaultExpireWindow,
			Pub:          pub,
		},
		UserPassProvider: userPassProvider,
		Client:           c,
	}
}

//Retrieve retrieves the token from the authentication server.
func (p *AuthAPIProvider) Retrieve() (string, error) {
	user, pass, err := p.UserPassProvider()
	if err != nil {
		return "", errors.Wrap(err, "failed to get username or password")
	}
	jwt, err := p.Client.GetToken(user, pass)
	if err != nil {
		if aerr, ok := err.(*v1payload.Error); ok {
			return "", aerr
		}
		return "", errors.Wrap(err, "failed to get nerd jwt for username and password")
	}
	err = p.SetExpirationFromJWT(jwt)
	if err != nil {
		return "", errors.Wrap(err, "failed to set expiration")
	}
	err = conf.WriteNerdToken(jwt)
	if err != nil {
		return "", errors.Wrap(err, "failed to write nerd jwt to config")
	}
	return jwt, nil
}
