package oauth

import (
	"fmt"
	"net/http"
	"time"

	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

//ErrTokenRevoked is returned when trying to refresh a revoked token
var ErrTokenRevoked = fmt.Errorf("ErrTokenRevoked")

//ErrTokenUnset is returned when no oatuh access token was found in the config file
var ErrTokenUnset = fmt.Errorf("ErrTokenUnset")

//ConfigProvider provides a oauth access token from the config file. For the default file location please see TokenFilename().
type ConfigProvider struct {
	*ProviderBasis
	Client *v1auth.OpsClient
}

//NewConfigProvider creates a new ConfigProvider provider.
func NewConfigProvider(client *v1auth.OpsClient) *ConfigProvider {
	return &ConfigProvider{
		ProviderBasis: &ProviderBasis{
			ExpireWindow: DefaultExpireWindow,
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
	if c.Credentials.OAuth.AccessToken == "" {
		return "", ErrTokenUnset
	}
	e.SetExpiration(c.Credentials.OAuth.Expiration)
	if e.IsExpired() {
		token, err := e.refresh(c.Credentials.OAuth.RefreshToken, c.Auth.ClientID)
		if err != nil {
			return "", errors.Wrap(err, "failed to refresh oauth access token")
		}
		return token, nil
	}
	return c.Credentials.OAuth.AccessToken, nil
}

//refresh refreshes the oath token with the refresh token
func (e *ConfigProvider) refresh(refreshToken, clientID string) (string, error) {
	out, err := e.Client.RefreshOAuthCredentials(refreshToken, clientID)
	if err != nil {
		if herr, ok := err.(*v1auth.HTTPError); ok && herr.StatusCode == http.StatusUnauthorized {
			return "", ErrTokenRevoked
		}
		return "", errors.Wrap(err, "failed to get oauth credentials")
	}
	expiration := time.Unix(e.CurrentTime().Unix()+int64(out.ExpiresIn), 0)
	e.SetExpiration(expiration)
	err = conf.WriteOAuth(out.AccessToken, out.RefreshToken, expiration, out.Scope, out.TokenType)
	if err != nil {
		return "", errors.Wrap(err, "failed to write oauth tokens to config")
	}
	return out.AccessToken, nil
}
