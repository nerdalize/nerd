package jwt

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

//DecodeTokenWithKey decodes a nerd token (JWT) and verifies it with the given public key.
func DecodeTokenWithKey(nerdToken string, key *ecdsa.PublicKey) (*jwt.StandardClaims, error) {
	return decodeToken(nerdToken, key)
}

//DecodeTokenWithPEM decodes a nerd token (JWT) and verifies it with the given public key in PEM format.
func DecodeTokenWithPEM(nerdToken, pem string) (*jwt.StandardClaims, error) {
	key, err := ParseECDSAPublicKeyFromPemBytes([]byte(pem))
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse public key PEM to ecdsa key")
	}
	return decodeToken(nerdToken, key)
}

//decodeToken decodes a nerd token (JWT) given the public key to check if the signature is valid.
func decodeToken(nerdToken string, key *ecdsa.PublicKey) (*jwt.StandardClaims, error) {
	p := &jwt.Parser{
		SkipClaimsValidation: true,
	}
	token, err := p.ParseWithClaims(nerdToken, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
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
	if claims, ok := token.Claims.(*jwt.StandardClaims); ok {
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

func isValid(jwt string, pub *ecdsa.PublicKey) error {
	claims, err := DecodeTokenWithKey(jwt, pub)
	if err != nil {
		return errors.Wrapf(err, "failed to decode jwt '%v'", jwt)
	}
	now := time.Now().Unix()
	if claims.VerifyExpiresAt(now, false) == false {
		delta := time.Unix(now, 0).Sub(time.Unix(claims.ExpiresAt, 0))
		return fmt.Errorf("token is expired by %v", delta)
	}
	return nil
}
