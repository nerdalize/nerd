package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"os"
	"strings"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var EmptyClaims = &jwt.StandardClaims{}

//testkey creates a new ecdsa keypair.
func testkey(t *testing.T) *ecdsa.PrivateKey {
	k, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	return k
}

//tokenAndPub returns a token string and public key.
func tokenAndPub(claims *jwt.StandardClaims, t *testing.T) (string, *ecdsa.PublicKey) {
	key := testkey(t)
	pub, ok := key.Public().(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("Could not cast ECDSA public key")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES384, claims)
	ss, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign claims: %v", err)
	}
	return ss, pub
}

//newEnvProvider creates a new Env provider
func newEnvProvider(pub *ecdsa.PublicKey) *EnvProvider {
	return &EnvProvider{
		ProviderBasis: &ProviderBasis{
			Pub: pub,
		},
	}
}

func TestEnvProviderRetrieve(t *testing.T) {
	// Success cases
	successCases := map[string]struct {
		claims *jwt.StandardClaims
	}{
		"valid": {
			claims: &jwt.StandardClaims{
				Audience: "nlz.com",
			},
		},
		"valid expired": {
			claims: &jwt.StandardClaims{
				Audience:  "nlz.com",
				ExpiresAt: time.Now().Unix() + 300,
			},
		},
	}
	for name, tc := range successCases {
		token, pub := tokenAndPub(tc.claims, t)
		if tc.claims == EmptyClaims {
			token = ""
		}
		os.Setenv("NERD_JWT", token)
		e := newEnvProvider(pub)
		value, err := e.Retrieve()
		if err != nil {
			t.Fatalf("%v: Unexpected error: %v", name, err)
		}
		if value != token {
			t.Errorf("%v: Expected token '%v', but got token '%v'", name, token, value)
		}
	}

	// Error cases
	errorCases := map[string]struct {
		claims   *jwt.StandardClaims
		errorMsg string
	}{
		"invalid expired": {
			claims: &jwt.StandardClaims{
				Audience:  "nlz.com",
				ExpiresAt: time.Now().Unix() - 300,
			},
			errorMsg: "is invalid",
		},
		"no token found": {
			claims:   EmptyClaims,
			errorMsg: "environment variable NERD_JWT is not set",
		},
	}
	for name, tc := range errorCases {
		token, pub := tokenAndPub(tc.claims, t)
		if tc.claims == EmptyClaims {
			token = ""
		}
		os.Setenv("NERD_JWT", token)

		e := newEnvProvider(pub)
		_, err := e.Retrieve()
		if err == nil {
			t.Fatalf("%v: Expected error but error was nil", name)
		}
		if !strings.Contains(err.Error(), tc.errorMsg) {
			t.Errorf("%v: Expected error '%v' message to contain '%v'", name, err, tc.errorMsg)
		}
	}
}
