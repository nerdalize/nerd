package credentials

import (
	"crypto/ecdsa"
	"crypto/subtle"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

const (
	NerdTokenEnvVar = "NERD_TOKEN"
	PublicKey       = `
-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEAkYbLnam4wo+heLlTZEeh1ZWsfruz9nk
kyvc4LwKZ8pez5KYY76H1ox+AfUlWOEq+bExypcFfEIrJkf/JXa7jpzkOWBDF9Sa
OWbQHMK+vvUXieCJvCc9Vj084ABwLBgX
-----END PUBLIC KEY-----
`
)

// TODO: When subject becomes string field in Thatcher's response, we can use the jwt.StandardClaims and remove all the validation code from this file. (https://github.com/nerdalize/authentication/issues/1)
type NerdClaims struct {
	Audience  string `json:"aud,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
	Id        string `json:"jti,omitempty"`
	IssuedAt  int64  `json:"iat,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	NotBefore int64  `json:"nbf,omitempty"`
	Subject   int64  `json:"sub,omitempty"`
	// *jwt.StandardClaims
}

func DecodeToken(nerdToken string) (*NerdClaims, error) {
	return DecodeTokenWithPEM(nerdToken, PublicKey)
}
func DecodeTokenWithPEM(nerdToken, pem string) (*NerdClaims, error) {
	key, err := ParseECDSAPublicKeyFromPemBytes([]byte(pem))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse public key PEM to ecdsa key")
	}
	return decodeToken(nerdToken, key)
}
func DecodeTokenWithKey(nerdToken string, key *ecdsa.PublicKey) (*NerdClaims, error) {
	return decodeToken(nerdToken, key)
}
func decodeToken(nerdToken string, key *ecdsa.PublicKey) (*NerdClaims, error) {
	// TODO: Prevent go-jwt from validating token while parsing
	token, err := jwt.ParseWithClaims(nerdToken, &NerdClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, errors.Errorf("Unexpected signing method: %v, expected ECDSA", token.Header["alg"])
		}

		return key, nil
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse nerd token '%v'", nerdToken)
	}
	if !token.Valid {
		return nil, errors.Errorf("nerd token '%v' signature is invalid", nerdToken)
	}
	if claims, ok := token.Claims.(*NerdClaims); ok {
		return claims, nil
	}

	return nil, errors.Errorf("could not decode nerd token '%v'", nerdToken)
}

//ParseECDSAPublicKeyFromPemBytes returns an ECDSA public key from pem bytes
func ParseECDSAPublicKeyFromPemBytes(pemb []byte) (*ecdsa.PublicKey, error) {
	block, _ := pem.Decode(pemb)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse")
	}

	switch pub := pub.(type) {
	case *ecdsa.PublicKey:
		return pub, nil
	default:
		return nil, errors.New("pem bytes doesn't contain a ECDSA public key")
	}
}

/////////////////////////////////////////////////////////////////
// Below code is not needed once https://github.com/nerdalize/authentication/issues/1 is fixed
////////////////////////////////////////////////////////////////

// Validates time based claims "exp, iat, nbf".
// There is no accounting for clock skew.
// As well, if any of the above claims are not in the token, it will still
// be considered a valid claim.
func (c NerdClaims) Valid() error {
	vErr := new(jwt.ValidationError)
	now := jwt.TimeFunc().Unix()

	// The claims below are optional, by default, so if they are set to the
	// default value in Go, let's not fail the verification for them.
	if c.VerifyExpiresAt(now, false) == false {
		delta := time.Unix(now, 0).Sub(time.Unix(c.ExpiresAt, 0))
		vErr.Inner = fmt.Errorf("token is expired by %v", delta)
		vErr.Errors |= jwt.ValidationErrorExpired
	}

	if c.VerifyIssuedAt(now, false) == false {
		vErr.Inner = fmt.Errorf("Token used before issued")
		vErr.Errors |= jwt.ValidationErrorIssuedAt
	}

	if c.VerifyNotBefore(now, false) == false {
		vErr.Inner = fmt.Errorf("token is not valid yet")
		vErr.Errors |= jwt.ValidationErrorNotValidYet
	}

	if vErr.Errors == 0 {
		return nil
	}

	return vErr
}

// Compares the aud claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (c *NerdClaims) VerifyAudience(cmp string, req bool) bool {
	return verifyAud(c.Audience, cmp, req)
}

// Compares the exp claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (c *NerdClaims) VerifyExpiresAt(cmp int64, req bool) bool {
	return verifyExp(c.ExpiresAt, cmp, req)
}

// Compares the iat claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (c *NerdClaims) VerifyIssuedAt(cmp int64, req bool) bool {
	return verifyIat(c.IssuedAt, cmp, req)
}

// Compares the iss claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (c *NerdClaims) VerifyIssuer(cmp string, req bool) bool {
	return verifyIss(c.Issuer, cmp, req)
}

// Compares the nbf claim against cmp.
// If required is false, this method will return true if the value matches or is unset
func (c *NerdClaims) VerifyNotBefore(cmp int64, req bool) bool {
	return verifyNbf(c.NotBefore, cmp, req)
}

// ----- helpers

func verifyAud(aud string, cmp string, required bool) bool {
	if aud == "" {
		return !required
	}
	if subtle.ConstantTimeCompare([]byte(aud), []byte(cmp)) != 0 {
		return true
	} else {
		return false
	}
}

func verifyExp(exp int64, now int64, required bool) bool {
	if exp == 0 {
		return !required
	}
	return now <= exp
}

func verifyIat(iat int64, now int64, required bool) bool {
	if iat == 0 {
		return !required
	}
	return now >= iat
}

func verifyIss(iss string, cmp string, required bool) bool {
	if iss == "" {
		return !required
	}
	if subtle.ConstantTimeCompare([]byte(iss), []byte(cmp)) != 0 {
		return true
	} else {
		return false
	}
}

func verifyNbf(nbf int64, now int64, required bool) bool {
	if nbf == 0 {
		return !required
	}
	return now >= nbf
}
