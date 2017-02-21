package credentials

import (
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type NerdClaims struct {
	Access    interface{} `json:"access"`
	Audience  string      `json:"aud"`
	ExpiresAt int64       `json:"exp,omitempty"`
	IssuedAt  int64       `json:"iat,omitempty"`
	Issuer    string      `json:"iss,omitempty"`
	NotBefore int64       `json:"nbf,omitempty"`
	Subject   string      `json:"sub,omitempty"`
}

// Validates time based claims "exp, iat, nbf".
// There is no accounting for clock skew.
// As well, if any of the above claims are not in the token, it will still
// be considered a valid claim.
func (c NerdClaims) Valid() error {
	now := time.Now().Unix()

	// The claims below are optional, by default, so if they are set to the
	// default value in Go, let's not fail the verification for them.
	if c.VerifyExpiresAt(now, false) == false {
		delta := time.Unix(now, 0).Sub(time.Unix(c.ExpiresAt, 0))
		return errors.Errorf("token is expired by %v", delta)
	}

	// if c.VerifyIssuedAt(now, false) == false {
	// 	return errors.Errorf("Token used before issued")
	// }
	//
	// if c.VerifyNotBefore(now, false) == false {
	// 	return errors.Errorf("token is not valid yet")
	// }

	return nil
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

// TODO: Use go-jwt package?
func DecodeToken(nerdToken string) (*NerdClaims, error) {
	// token, err := jwt.ParseWithClaims(nerdToken, NerdClaims{}, func(token *jwt.Token) (interface{}, error) {
	// 	// Don't forget to validate the alg is what you expect:
	// 	// if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
	// 	// 	return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	// 	// }
	//
	// 	// TODO: include public key?
	// 	return nil, nil
	// })
	// if err != nil {
	// 	return nil, errors.Wrapf(err, "failed to parse nerd token '%v'", nerdToken)
	// }
	//
	// //TODO Check if token is valid?
	// // if token.Valid {
	// if claims, ok := token.Claims.(*NerdClaims); ok {
	// 	return claims, nil
	// }
	//
	// return nil, errors.Errorf("could not decode nerd token '%v'", nerdToken)
	split := strings.Split(nerdToken, ".")
	if len(split) != 3 {
		return nil, errors.Errorf("token '%v' should consist of three parts", nerdToken)
	}
	dec, err := DecodeSegment(split[1])
	if err != nil {
		return nil, errors.Wrapf(err, "token '%v' payload could not be base64 decoded", nerdToken)
	}
	res := &NerdClaims{}
	err = json.Unmarshal(dec, res)
	if err != nil {
		return nil, errors.Wrapf(err, "token '%v' payload (%v) could not be json decoded", nerdToken, string(dec))
	}
	return res, nil
}

func DecodeSegment(seg string) ([]byte, error) {
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}

	return base64.URLEncoding.DecodeString(seg)
}
