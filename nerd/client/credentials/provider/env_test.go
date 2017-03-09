package provider

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"os"
	"strings"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/nerdalize/nerd/nerd/payload"
)

var EmptyClaims = &payload.NerdClaims{}

//testkey creates a new ecdsa keypair.
func testkey(t *testing.T) *ecdsa.PrivateKey {
	k, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	return k
}

//tokenAndPub returns a token string and public key.
func tokenAndPub(claims *payload.NerdClaims, t *testing.T) (string, *ecdsa.PublicKey) {
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
func newEnvProvider(pub *ecdsa.PublicKey) *Env {
	return &Env{
		Basis: &Basis{},
	}
}

func TestEnvProviderRetrieve(t *testing.T) {
	// Success cases
	successCases := map[string]struct {
		claims *payload.NerdClaims
	}{
		"valid": {
			claims: &payload.NerdClaims{
				StandardClaims: &jwt.StandardClaims{
					Audience: "nlz.com",
				},
			},
		},
		"valid expired": {
			claims: &payload.NerdClaims{
				StandardClaims: &jwt.StandardClaims{
					Audience:  "nlz.com",
					ExpiresAt: time.Now().Unix() + 300,
				},
			},
		},
	}
	for name, tc := range successCases {
		token, pub := tokenAndPub(tc.claims, t)
		if tc.claims == EmptyClaims {
			token = ""
		}
		os.Setenv("NERD_TOKEN", token)
		e := newEnvProvider(pub)
		value, err := e.Retrieve(pub)
		if err != nil {
			t.Fatalf("%v: Unexpected error: %v", name, err)
		}
		if value.NerdToken != token {
			t.Errorf("%v: Expected token '%v', but got token '%v'", name, token, value.NerdToken)
		}
	}

	// Error cases
	errorCases := map[string]struct {
		claims   *payload.NerdClaims
		errorMsg string
	}{
		"invalid expired": {
			claims: &payload.NerdClaims{
				StandardClaims: &jwt.StandardClaims{
					Audience:  "nlz.com",
					ExpiresAt: time.Now().Unix() - 300,
				},
			},
			errorMsg: "is invalid",
		},
		"no token found": {
			claims:   EmptyClaims,
			errorMsg: "environment variable NERD_TOKEN is not set",
		},
	}
	for name, tc := range errorCases {
		token, pub := tokenAndPub(tc.claims, t)
		if tc.claims == EmptyClaims {
			token = ""
		}
		os.Setenv("NERD_TOKEN", token)

		e := newEnvProvider(pub)
		_, err := e.Retrieve(pub)
		if err == nil {
			t.Fatalf("%v: Expected error but error was nil", name)
		}
		if !strings.Contains(err.Error(), tc.errorMsg) {
			t.Errorf("%v: Expected error '%v' message to contain '%v'", name, err, tc.errorMsg)
		}
	}
}
