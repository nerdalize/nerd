package provider

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/nerdalize/nerd/nerd/client/credentials"
)

var EmptyClaims = &credentials.NerdClaims{}

func testkey(t *testing.T) *ecdsa.PrivateKey {
	k, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	return k
}

func tokenAndPub(claims *credentials.NerdClaims, t *testing.T) (string, *ecdsa.PublicKey) {
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

func TestEnvProviderRetrieve(t *testing.T) {
	testCases := map[string]struct {
		claims     *credentials.NerdClaims
		success    bool
		createFile bool
		errorMsg   string
	}{
		"valid": {
			claims: &credentials.NerdClaims{
				StandardClaims: &jwt.StandardClaims{
					Audience: "nlz.com",
				},
			},
			success: true,
		},
		"valid expired": {
			claims: &credentials.NerdClaims{
				StandardClaims: &jwt.StandardClaims{
					Audience:  "nlz.com",
					ExpiresAt: time.Now().Unix() + 300,
				},
			},
			success: true,
		},
		"invalid expired": {
			claims: &credentials.NerdClaims{
				StandardClaims: &jwt.StandardClaims{
					Audience:  "nlz.com",
					ExpiresAt: time.Now().Unix() - 300,
				},
			},
			success:  false,
			errorMsg: "is invalid",
		},
		"no token found": {
			claims:   EmptyClaims,
			success:  false,
			errorMsg: "are empty",
		},
		"valid file": {
			claims: &credentials.NerdClaims{
				StandardClaims: &jwt.StandardClaims{
					Audience: "nlz.com",
				},
			},
			success:    true,
			createFile: true,
		},
	}
	for name, tc := range testCases {
		token, pub := tokenAndPub(tc.claims, t)
		if tc.claims == EmptyClaims {
			token = ""
		}
		temp, err := ioutil.TempFile("/tmp", "token")
		defer temp.Close()
		if err != nil {
			t.Fatalf("%v: Unexpected error for temp file: %v", name, err)
		}
		if tc.createFile {
			temp.WriteString(token)
		} else {
			os.Setenv("NERD_TOKEN", token)
		}

		e := &EnvDisk{
			TokenDecoder: func(t string) (*credentials.NerdClaims, error) {
				return credentials.DecodeTokenWithKey(t, pub)
			},
			DiskLocation: temp.Name(),
		}
		value, err := e.Retrieve()
		if tc.success {
			if err != nil {
				t.Fatalf("%v: Unexpected error: %v", name, err)
			}
			if value.NerdToken != token {
				t.Errorf("%v: Expected token '%v', but got token '%v'", name, token, value.NerdToken)
			}
		} else {
			if err == nil {
				t.Fatalf("%v: Expected error but error was nil", name)
			}
			if !strings.Contains(err.Error(), tc.errorMsg) {
				t.Errorf("%v: Expected error '%v' message to contain '%v'", name, err, tc.errorMsg)
			}
		}
	}
}
