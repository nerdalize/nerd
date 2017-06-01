package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	v1payload "github.com/nerdalize/nerd/nerd/client/auth/v1/payload"
)

const (
	minute = 60
)

//testkey creates a new ecdsa keypair.
func testkey(t *testing.T) *ecdsa.PrivateKey {
	k, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	return k
}

//tokenAndPub returns a token string and public key.
func getToken(key *ecdsa.PrivateKey, claims *jwt.StandardClaims, t *testing.T) string {
	token := jwt.NewWithClaims(jwt.SigningMethodES384, claims)
	ss, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign claims: %v", err)
	}
	return ss
}

type tokenClient struct {
	token string
}

//RefreshJWT refreshes a JWT with a refresh token
func (c *tokenClient) RefreshJWT(projectID, jwt, secret string) (output *v1payload.RefreshWorkerJWTOutput, err error) {
	output = &v1payload.RefreshWorkerJWTOutput{
		Token: c.token,
	}
	return output, nil
}

//RevokeJWT revokes a JWT
func (c *tokenClient) RevokeJWT(projectID, jwt, secret string) (output *v1payload.RefreshWorkerJWTOutput, err error) {
	return nil, fmt.Errorf("not implemented")
}

func timeFunc(t time.Time) func() time.Time {
	return func() time.Time {
		return t
	}
}
