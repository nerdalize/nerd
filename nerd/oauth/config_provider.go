package oauth

import (
	"net/http"
	"time"

	"github.com/nerdalize/nerd/nerd"
	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

//ConfigProvider provides a oauth access token from the config file. For the default file location please see TokenFilename().
type ConfigProvider struct {
	*ProviderBasis
	Client        v1auth.OpsClientInterface
	OAuthClientID string
	Session       conf.SessionInterface
}

//NewConfigProvider creates a new ConfigProvider provider.
func NewConfigProvider(client v1auth.OpsClientInterface, oauthClientID string, session conf.SessionInterface) *ConfigProvider {
	return &ConfigProvider{
		ProviderBasis: &ProviderBasis{
			ExpireWindow: DefaultExpireWindow,
		},
		Client:        client,
		OAuthClientID: oauthClientID,
		Session:       session,
	}
}

//Retrieve retrieves the token from the nerd config file.
func (e *ConfigProvider) Retrieve() (string, error) {
	ss, err := e.Session.Read()
	if err != nil {
		return "", errors.Wrap(err, "failed to read config")
	}
	if ss.OAuth.AccessToken == "" {
		return "", nerd.ErrTokenUnset
	}
	e.SetExpiration(ss.OAuth.Expiration)
	if e.IsExpired() {
		token, err := e.refresh(ss.OAuth.RefreshToken, e.OAuthClientID)
		if err != nil {
			return "", errors.Wrap(err, "failed to refresh oauth access token")
		}
		return token, nil
	}
	return ss.OAuth.AccessToken, nil
}

//refresh refreshes the oath token with the refresh token
func (e *ConfigProvider) refresh(refreshToken, clientID string) (string, error) {
	out, err := e.Client.RefreshOAuthCredentials(refreshToken, clientID)
	if err != nil {
		if herr, ok := err.(*v1auth.HTTPError); ok && herr.StatusCode == http.StatusUnauthorized {
			return "", nerd.ErrTokenRevoked
		}
		return "", errors.Wrap(err, "failed to get oauth credentials")
	}
	expiration := time.Unix(e.CurrentTime().Unix()+int64(out.ExpiresIn), 0)
	e.SetExpiration(expiration)
	err = e.Session.WriteOAuth(out.AccessToken, out.RefreshToken, "", expiration, out.Scope, out.TokenType)
	if err != nil {
		return "", errors.Wrap(err, "failed to write oauth tokens to config")
	}
	return out.AccessToken, nil
}
