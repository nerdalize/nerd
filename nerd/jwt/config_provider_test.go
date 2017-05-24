package jwt

import (
	"crypto/ecdsa"
	"fmt"
	"strings"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/nerd/utils"
)

type inMemSession struct {
	ss *conf.SessionSnapshot
}

//Read returns a snapshot of the session file
func (s *inMemSession) Read() (*conf.SessionSnapshot, error) {
	return s.ss, nil
}

//WriteJWT writes the jwt object to the session file
func (s *inMemSession) WriteJWT(jwt, refreshToken string) error {
	s.ss.JWT.Token = jwt
	s.ss.JWT.RefreshToken = refreshToken
	return nil
}

//WriteOAuth writes the oauth object to the session file
func (s *inMemSession) WriteOAuth(accessToken, refreshToken string, expiration time.Time, scope, tokenType string) error {
	return fmt.Errorf("not implemented")
}

//WriteProject writes the project object to the session file
func (s *inMemSession) WriteProject(name, awsRegion string) error {
	return fmt.Errorf("not implemented")
}

func TestConfigProvider(t *testing.T) {
	key := testkey(t)
	pub, _ := key.Public().(*ecdsa.PublicKey)
	now := time.Now().Unix()
	refreshedClaims := &jwt.StandardClaims{
		ExpiresAt: now + minute*10,
	}
	refreshedToken := getToken(key, refreshedClaims, t)
	session := &inMemSession{
		ss: &conf.SessionSnapshot{},
	}
	client := &tokenClient{
		token: refreshedToken,
	}

	prov := NewConfigProvider(pub, session, client)
	prov.ExpireWindow = 0

	t.Run("normal", func(t *testing.T) {
		claims := &jwt.StandardClaims{
			ExpiresAt: now + minute*5,
		}
		token := getToken(key, claims, t)
		session.WriteJWT(token, "")
		ret, err := prov.Retrieve()
		utils.OK(t, err)
		utils.Equals(t, token, ret)
	})

	t.Run("noToken", func(t *testing.T) {
		session.WriteJWT("", "")
		ret, err := prov.Retrieve()
		utils.Assert(t, err != nil, "expected error because no token was set")
		utils.Assert(t, strings.Contains(err.Error(), "not set"), "expected error because no token was set", err)
		utils.Equals(t, "", ret)
	})

	claimsExp := &jwt.StandardClaims{
		ExpiresAt: now - minute*5,
	}
	tokenExp := getToken(key, claimsExp, t)
	t.Run("expired", func(t *testing.T) {
		session.WriteJWT(tokenExp, "")
		ret, err := prov.Retrieve()
		utils.Assert(t, err != nil, "expected token to be expired")
		utils.Assert(t, strings.Contains(err.Error(), "expired"), "expected token to be expired", err)
		utils.Equals(t, "", ret)
	})

	t.Run("refresh", func(t *testing.T) {
		secret := "abc"
		session.WriteJWT(tokenExp, secret)
		ret, err := prov.Retrieve()
		utils.OK(t, err)
		utils.Equals(t, refreshedToken, ret)
		utils.Equals(t, refreshedClaims.ExpiresAt, prov.expiration.Unix())
		ss, _ := session.Read()
		utils.Equals(t, refreshedToken, ss.JWT.Token)
		utils.Equals(t, secret, ss.JWT.RefreshToken)
	})
}
