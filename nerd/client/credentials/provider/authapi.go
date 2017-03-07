package provider

import (
	"crypto/ecdsa"
	"io/ioutil"
	"time"

	"github.com/nerdalize/nerd/nerd/client"
	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

const NerdTokenPermissions = 0644

//AuthAPI provides nerdalize credentials by making a request to the nerdalize auth server.
//The UserPassProvider is used to retreive the username and password required to authenticate with the auth server.
type AuthAPI struct {
	*ProviderBasis

	Client           *client.AuthAPIClient
	UserPassProvider func() (string, string, error)
}

func NewAuthAPI(userPassProvider func() (string, string, error), c *client.AuthAPIClient) *AuthAPI {
	return &AuthAPI{
		ProviderBasis: &ProviderBasis{
			ExpireWindow: DefaultExpireWindow,
		},
		UserPassProvider: userPassProvider,
		Client:           c,
	}
}

func (p *AuthAPI) Retrieve(pub *ecdsa.PublicKey) (*credentials.NerdAPIValue, error) {
	user, pass, err := p.UserPassProvider()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get username or password")
	}
	token, err := p.Client.GetToken(user, pass)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get nerd token for username and password")
	}
	err = saveNerdToken(token)
	if err != nil {
		return nil, errors.Wrap(err, "failed to save nerd token")
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

//saveNerdToken saves the token to disk.
func saveNerdToken(token string) error {
	filename, err := TokenFilename()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, []byte(token), NerdTokenPermissions)
	if err != nil {
		return errors.Wrapf(err, "failed to write nerd token to '%v'", filename)
	}
	return nil
}
