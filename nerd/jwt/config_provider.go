package jwt

import (
	"crypto/ecdsa"
	"time"

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
	jwt := c.NerdToken
	if jwt == "" {
		return "", errors.New("nerd_token is not set in config")
	}
	claims, err := DecodeTokenWithKey(jwt, e.Pub)
	if err != nil {
		return "", errors.Wrapf(err, "failed to decode jwt '%v'", jwt)
	}
	e.AlwaysValid = claims.ExpiresAt == 0 // if unset
	e.SetExpiration(time.Unix(claims.ExpiresAt, 0))
	err = claims.Valid()
	if err != nil {
		return "", errors.Wrapf(err, "nerd jwt '%v' is invalid", jwt)
	}
	return jwt, nil
}
