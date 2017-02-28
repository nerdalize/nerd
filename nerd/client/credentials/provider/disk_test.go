package provider

import (
	"crypto/ecdsa"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/nerdalize/nerd/nerd/client/credentials"
)

func newDiskProvider(pub *ecdsa.PublicKey, token string, t *testing.T) *Disk {
	temp, err := ioutil.TempFile("/tmp", "token")
	if err != nil {
		t.Fatalf("Unexpected error for temp file: %v", err)
	}
	temp.WriteString(token)
	return &Disk{
		DiskLocation: temp.Name(),
		ProviderBasis: &ProviderBasis{
			TokenDecoder: func(t string) (*credentials.NerdClaims, error) {
				return credentials.DecodeTokenWithKey(t, pub)
			},
		},
	}
}

func TestDiskProviderRetrieve(t *testing.T) {
	successCases := map[string]struct {
		claims *credentials.NerdClaims
	}{
		"valid": {
			claims: &credentials.NerdClaims{
				StandardClaims: &jwt.StandardClaims{
					Audience: "nlz.com",
				},
			},
		},
		"valid expired": {
			claims: &credentials.NerdClaims{
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
		d := newDiskProvider(pub, token, t)
		value, err := d.Retrieve()
		if err != nil {
			t.Fatalf("%v: Unexpected error: %v", name, err)
		}
		if value.NerdToken != token {
			t.Errorf("%v: Expected token '%v', but got token '%v'", name, token, value.NerdToken)
		}
	}
	errorCases := map[string]struct {
		claims   *credentials.NerdClaims
		errorMsg string
	}{
		"invalid expired": {
			claims: &credentials.NerdClaims{
				StandardClaims: &jwt.StandardClaims{
					Audience:  "nlz.com",
					ExpiresAt: time.Now().Unix() - 300,
				},
			},
			errorMsg: "is invalid",
		},
		"no token found": {
			claims:   EmptyClaims,
			errorMsg: "is empty",
		},
	}
	for name, tc := range errorCases {
		token, pub := tokenAndPub(tc.claims, t)
		if tc.claims == EmptyClaims {
			token = ""
		}
		d := newDiskProvider(pub, token, t)
		_, err := d.Retrieve()
		if err == nil {
			t.Fatalf("%v: Expected error but error was nil", name)
		}
		if !strings.Contains(err.Error(), tc.errorMsg) {
			t.Errorf("%v: Expected error '%v' message to contain '%v'", name, err, tc.errorMsg)
		}
	}
}
