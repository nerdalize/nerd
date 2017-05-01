package jwt

import (
	"crypto/ecdsa"

	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

//ConfigProvider provides a JWT from the config file. For the default file location please see TokenFilename().
type ConfigProvider struct {
	*ProviderBasis
}

//NewConfigProvider creates a new ConfigProvider provider.
func NewConfigProvider(pub *ecdsa.PublicKey) *ConfigProvider {
	return &ConfigProvider{
		ProviderBasis: &ProviderBasis{
			ExpireWindow: DefaultExpireWindow,
			Pub:          pub,
		},
	}
}

//Retrieve retrieves the token from the nerd config file.
func (e *ConfigProvider) Retrieve() (string, error) {
	c, err := conf.Read()
	if err != nil {
		return "", errors.Wrap(err, "failed to read config")
	}
	if c.NerdToken == "" {
		return "", errors.New("nerd_token is not set in config")
	}
	err = e.SetExpirationFromJWT(c.NerdToken)
	if err != nil {
		return "", errors.Wrap(err, "failed to set expiration")
	}
	return c.NerdToken, nil
}
