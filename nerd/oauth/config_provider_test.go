package oauth

import (
	"fmt"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	v1payload "github.com/nerdalize/nerd/nerd/client/auth/v1/payload"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/nerd/utils"
)

const (
	minute = 60
)

var EmptyClaims = &jwt.StandardClaims{}

type inMemSession struct {
	ss *conf.SessionSnapshot
}

//Read returns a snapshot of the session file
func (s *inMemSession) Read() (*conf.SessionSnapshot, error) {
	return s.ss, nil
}

//WriteJWT writes the jwt object to the session file
func (s *inMemSession) WriteJWT(jwt, refreshToken string) error {
	return fmt.Errorf("not implemented")
}

type opsClient struct {
	output *v1payload.RefreshOAuthCredentialsOutput
}

//GetOAuthCredentials gets oauth credentials based on a 'session' code
func (c *opsClient) GetOAuthCredentials(code, clientID, localServerURL string) (output *v1payload.GetOAuthCredentialsOutput, err error) {
	return nil, fmt.Errorf("not implemented")
}

//RefreshOAuthCredentials refreshes an oauth access token
func (c *opsClient) RefreshOAuthCredentials(refreshToken, clientID string) (output *v1payload.RefreshOAuthCredentialsOutput, err error) {
	return c.output, nil
}

//WriteOAuth writes the oauth object to the session file
func (s *inMemSession) WriteOAuth(accessToken, refreshToken string, expiration time.Time, scope, tokenType string) error {
	s.ss.OAuth.AccessToken = accessToken
	s.ss.OAuth.RefreshToken = refreshToken
	s.ss.OAuth.Expiration = expiration
	s.ss.OAuth.Scope = scope
	s.ss.OAuth.TokenType = tokenType
	return nil
}

//WriteProject writes the project object to the session file
func (s *inMemSession) WriteProject(name, awsRegion string) error {
	return fmt.Errorf("not implemented")
}

func TestConfigProvider(t *testing.T) {
	session := &inMemSession{
		ss: &conf.SessionSnapshot{},
	}
	refreshedCredentials := &v1payload.RefreshOAuthCredentialsOutput{
		OAuthCredentials: v1payload.OAuthCredentials{
			AccessToken:  "new_token",
			RefreshToken: "new_refreshed_token",
			ExpiresIn:    5 * minute,
		},
	}
	client := &opsClient{
		output: refreshedCredentials,
	}

	prov := NewConfigProvider(client, "", session)
	prov.ExpireWindow = 0

	t.Run("normal", func(t *testing.T) {
		token := "default"
		session.WriteOAuth(token, "normal_refresh", time.Now().Add(5*time.Minute), "", "")
		ret, err := prov.Retrieve()
		utils.OK(t, err)
		utils.Equals(t, token, ret)
	})

	t.Run("expired", func(t *testing.T) {
		exp := time.Now().Add(-5 * time.Minute)
		session.WriteOAuth("default", "normal_refresh", exp, "", "")
		ret, err := prov.Retrieve()
		utils.OK(t, err)
		utils.Equals(t, refreshedCredentials.AccessToken, ret)
		utils.Assert(t, prov.expiration.Unix() > exp.Unix(), "expected new expiration date to be set", prov.expiration, exp)
	})

}
