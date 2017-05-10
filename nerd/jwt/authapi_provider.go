package jwt

import (
	"crypto/ecdsa"

	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

const (
	jwtScope = "nce.nerdalize.com"
)

//AuthAPIProvider provides nerdalize credentials by making a request to the nerdalize auth server.
//The UserPassProvider is used to retreive the username and password required to authenticate with the auth server.
type AuthAPIProvider struct {
	*ProviderBasis

	Client *v1auth.Client
}

//NewAuthAPIProvider creates a new AuthAPIProvider provider.
func NewAuthAPIProvider(pub *ecdsa.PublicKey, c *v1auth.Client) *AuthAPIProvider {
	return &AuthAPIProvider{
		ProviderBasis: &ProviderBasis{
			ExpireWindow: DefaultExpireWindow,
			Pub:          pub,
		},
		Client: c,
	}
}

//Retrieve retrieves the token from the authentication server.
func (p *AuthAPIProvider) Retrieve() (string, error) {
	out, err := p.Client.GetJWT(jwtScope)
	if err != nil {
		return "", errors.Wrap(err, "failed to get nerd jwt")
	}
	err = p.SetExpirationFromJWT(out.Token)
	if err != nil {
		return "", errors.Wrap(err, "failed to set expiration")
	}
	err = conf.WriteJWT(out.Token, "")
	if err != nil {
		return "", errors.Wrap(err, "failed to write nerd jwt to config")
	}
	return out.Token, nil
}
