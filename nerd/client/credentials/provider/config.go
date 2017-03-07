package provider

import (
	"crypto/ecdsa"
	"time"

	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

//Config provides nerdalize credentials from a file on Config. The default file can be found in TokenFilename().
type Config struct {
	*ProviderBasis
}

func NewConfigCredentials(pub *ecdsa.PublicKey) *credentials.NerdAPI {
	return credentials.NewNerdAPI(pub, NewConfig())
}

func NewConfig() *Config {
	return &Config{
		ProviderBasis: &ProviderBasis{
			ExpireWindow: DefaultExpireWindow,
		},
	}
}

func (e *Config) Retrieve(pub *ecdsa.PublicKey) (*credentials.NerdAPIValue, error) {
	c, err := conf.Read()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config")
	}
	token := c.NerdToken
	if token == "" {
		return nil, errors.New("nerd_token is not set in config")
	}
	claims, err := credentials.DecodeTokenWithKey(token, pub)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode token '%v'", token)
	}
	e.AlwaysValid = claims.ExpiresAt == 0 // if unset
	e.SetExpiration(time.Unix(claims.ExpiresAt, 0))
	err = claims.Valid()
	if err != nil {
		return nil, errors.Wrapf(err, "nerd token '%v' is invalid", token)
	}
	return &credentials.NerdAPIValue{NerdToken: token}, nil
}
