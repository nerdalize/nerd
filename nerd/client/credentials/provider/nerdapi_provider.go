package provider

import (
	"io/ioutil"
	"time"

	"github.com/nerdalize/nerd/nerd/client"
	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/pkg/errors"
)

const NerdTokenPermissions = 0644
const DefaultExpireWindow = 20

type NerdAPIProvider struct {
	// The date/time when to expire on
	expiration time.Time

	// If set will be used by IsExpired to determine the current time.
	// Defaults to time.Now if CurrentTime is not set.  Available for testing
	// to be able to mock out the current time.
	CurrentTime func() time.Time

	Client       *client.AuthAPIClient
	ExpireWindow time.Duration

	UserPassProvider func() (string, string, error)
}

func NewNerdAPIProvider(userPassProvider func() (string, string, error), c *client.AuthAPIClient) *NerdAPIProvider {
	return &NerdAPIProvider{
		ExpireWindow:     DefaultExpireWindow,
		UserPassProvider: userPassProvider,
		Client:           c,
	}
}

// IsExpired returns true if the credentials retrieved are expired, or not yet
// retrieved.
// TODO: Test expired things, also include exp in JWT
func (p *NerdAPIProvider) IsExpired() bool {
	if p.CurrentTime == nil {
		p.CurrentTime = time.Now
	}
	return p.expiration.Before(p.CurrentTime())
}

func (p *NerdAPIProvider) SetExpiration(expiration time.Time) {
	p.expiration = expiration
	if p.ExpireWindow > 0 {
		p.expiration = p.expiration.Add(-p.ExpireWindow)
	}
}

func saveNerdToken(token string) error {
	filename, err := credentials.TokenFilename()
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, []byte(token), NerdTokenPermissions)
	if err != nil {
		return errors.Wrapf(err, "failed to write nerd token to '%v'", filename)
	}
	return nil
}

// Retrieve will attempt to request the credentials from the endpoint the Provider
// was configured for. And error will be returned if the retrieval fails.
func (p *NerdAPIProvider) Retrieve() (*credentials.NerdAPIValue, error) {
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
	claims, err := credentials.DecodeToken(token)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retreive claims from nerd token '%v'", token)
	}
	p.SetExpiration(time.Unix(claims.ExpiresAt, 0))
	return &credentials.NerdAPIValue{
		NerdToken: token,
	}, nil
}
