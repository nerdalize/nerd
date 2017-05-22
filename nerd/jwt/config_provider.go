package jwt

import (
	"crypto/ecdsa"

	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

//ConfigProvider provides a JWT from the config file. For the default file location please see TokenFilename().
type ConfigProvider struct {
	*ProviderBasis
	Session conf.SessionInterface
}

//NewConfigProvider creates a new ConfigProvider provider.
func NewConfigProvider(pub *ecdsa.PublicKey, session conf.SessionInterface) *ConfigProvider {
	return &ConfigProvider{
		ProviderBasis: &ProviderBasis{
			ExpireWindow: DefaultExpireWindow,
			Pub:          pub,
		},
		Session: session,
	}
}

//Retrieve retrieves the token from the nerd config file.
func (e *ConfigProvider) Retrieve() (string, error) {
	c, err := e.Session.Read()
	if err != nil {
		return "", errors.Wrap(err, "failed to read config")
	}
	if c.JWT.Token == "" {
		return "", errors.New("nerd_token is not set in config")
	}
	err = e.SetExpirationFromJWT(c.JWT.Token)
	if err != nil {
		return "", errors.Wrap(err, "failed to set expiration")
	}
	return c.JWT.Token, nil
}
