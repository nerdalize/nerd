package credentials

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"strings"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func TestParseECDSAPublicKeyFromPemBytes(t *testing.T) {
	key := `
-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEAkYbLnam4wo+heLlTZEeh1ZWsfruz9nk
kyvc4LwKZ8pez5KYY76H1ox+AfUlWOEq+bExypcFfEIrJkf/JXa7jpzkOWBDF9Sa
OWbQHMK+vvUXieCJvCc9Vj084ABwLBgX
-----END PUBLIC KEY----
`
	_, err := ParseECDSAPublicKeyFromPemBytes([]byte(key))
	if err != nil {
		t.Errorf("Failed to parse valid public key. Error message: %v", err)
	}
}

func testkey(tb testing.TB) *ecdsa.PrivateKey {
	k, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		tb.Fatalf("failed to generate test key: %v", err)
	}
	return k
}

func TestDecodeToken(t *testing.T) {
	key := testkey(t)
	pub, ok := key.Public().(*ecdsa.PublicKey)
	if !ok {
		t.Fatal("Could not cast ECDSA public key")
	}
	successTime := time.Now().Add(time.Hour).Unix()
	testCases := map[string]struct {
		claims    *NerdClaims
		token     string
		success   bool
		errorMsg  string
		expiresAt int64
		audience  string
	}{
		"success": {
			claims: &NerdClaims{
				Audience:  "nlz.com",
				ExpiresAt: successTime,
			},
			token:     "",
			success:   true,
			errorMsg:  "",
			expiresAt: successTime,
			audience:  "nlz.com",
		},
		"json parse error": {
			claims:    &NerdClaims{},
			token:     "jwt.jwt.jwt",
			success:   false,
			errorMsg:  "failed to parse nerd token",
			expiresAt: 111,
			audience:  "nlz.com",
		},
	}

	for name, tc := range testCases {
		token := jwt.NewWithClaims(jwt.SigningMethodES384, tc.claims)
		ss, err := token.SignedString(key)
		if err != nil {
			t.Fatalf("failed to sign test token: %v", err)
		}
		if tc.token != "" {
			ss = tc.token
		}
		claims, err := DecodeTokenWithKey(ss, pub)
		if tc.success {
			if err != nil {
				t.Errorf("%v: expected success but got error '%v'", name, err)
				continue
			}
			if claims.Audience != tc.audience {
				t.Errorf("%v: expected audience to be '%v' but was '%v'", name, tc.audience, claims.Audience)
			}
			if claims.ExpiresAt != tc.expiresAt {
				t.Errorf("%v: expected expiresAt to be '%v' but was '%v'", name, tc.expiresAt, claims.ExpiresAt)
			}
		} else {
			if err == nil {
				t.Errorf("%v: expected failure but got success", name)
				continue
			}
			if !strings.Contains(err.Error(), tc.errorMsg) {
				t.Errorf("%v: expected error message to contain '%v' but error message was '%v'", name, tc.errorMsg, err.Error())
			}
		}
	}
}
