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
	Session conf.SessionInterface
	Client  v1auth.TokenClientInterface
}

//NewConfigProvider creates a new ConfigProvider provider.
func NewConfigProvider(pub *ecdsa.PublicKey, session conf.SessionInterface, client v1auth.TokenClientInterface) *ConfigProvider {
	return &ConfigProvider{
		ProviderBasis: &ProviderBasis{
			ExpireWindow: DefaultExpireWindow,
			Pub:          pub,
		},
		Session: session,
		Client:  client,
	}
}

//Retrieve retrieves the token from the nerd config file.
func (e *ConfigProvider) Retrieve() (string, error) {
	ss, err := e.Session.Read()
	if err != nil {
		return "", errors.Wrap(err, "failed to read config")
	}
	jwt := ss.JWT.Token
	if jwt == "" {
		return "", errors.New(".jwt.token is not set in config")
	}
	err = e.SetExpirationFromJWT(jwt)
	if err != nil {
		return "", errors.Wrap(err, "failed to set expiration")
	}
	if ss.JWT.RefreshToken != "" && e.IsExpired() {
		jwt, err = e.refresh(jwt, ss.JWT.RefreshToken, ss.Project.Name)
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

func (e *ConfigProvider) refresh(jwt, secret, projectID string) (string, error) {
	out, err := e.Client.RefreshJWT(projectID, jwt, secret)
	if err != nil {
		return "", errors.Wrap(err, "failed to refresh token")
	}
	err = e.SetExpirationFromJWT(out.Token)
	if err != nil {
		return "", errors.Wrap(err, "failed to set expiration")
	}
	err = e.Session.WriteJWT(out.Token, secret)
	if err != nil {
		return "", errors.Wrap(err, "failed to write jwt and secret to config")
	}
	return out.Token, nil
}
