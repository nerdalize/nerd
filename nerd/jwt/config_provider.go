package jwt

import (
	"crypto/ecdsa"

	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

//ConfigProvider provides a JWT from the config file. For the default file location please see TokenFilename().
type ConfigProvider struct {
	*ProviderBasis
	Client *v1auth.TokenClient
}

//NewConfigProvider creates a new ConfigProvider provider.
func NewConfigProvider(pub *ecdsa.PublicKey, client *v1auth.TokenClient) *ConfigProvider {
	return &ConfigProvider{
		ProviderBasis: &ProviderBasis{
			ExpireWindow: DefaultExpireWindow,
			Pub:          pub,
		},
		Client: client,
	}
}

//Retrieve retrieves the token from the nerd config file.
func (e *ConfigProvider) Retrieve() (string, error) {
	c, err := conf.Read()
	if err != nil {
		return "", errors.Wrap(err, "failed to read config")
	}
	jwt := c.Credentials.JWT.Token
	if jwt == "" {
		return "", errors.New("nerd_token is not set in config")
	}
	err = e.SetExpirationFromJWT(jwt)
	if err != nil {
		return "", errors.Wrap(err, "failed to set expiration")
	}
	if c.Credentials.JWT.RefreshToken != "" && e.IsExpired() {
		jwt, err = e.refresh(jwt, c.Credentials.JWT.RefreshToken)
		if err != nil {
			return "", errors.Wrap(err, "failed to refresh")
		}
	}
	err = isValid(jwt, e.Pub)
	if err != nil {
		return "", err
	}
	return jwt, nil
}

func (e *ConfigProvider) refresh(jwt, secret string) (string, error) {
	config, err := conf.Read()
	if err != nil {
		return "", errors.Wrap(err, "failed to read config")
	}
	out, err := e.Client.RefreshJWT(config.CurrentProject.Name, jwt, secret)
	if err != nil {
		return "", errors.Wrap(err, "failed to refresh token")
	}
	err = e.SetExpirationFromJWT(out.Token)
	if err != nil {
		return "", errors.Wrap(err, "failed to set expiration")
	}
	err = conf.WriteJWT(out.Token, out.Secret)
	if err != nil {
		return "", errors.Wrap(err, "failed to write jwt and secret to config")
	}
	return out.Token, nil
}
