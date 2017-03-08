package credentials

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

const (
	NerdTokenEnvVar = "NERD_TOKEN"
	PublicKey       = `-----BEGIN PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEq+vArxUax61kU9w3i27s5xQQt0Qdlm58
hNnFGSq5sktUgv9UcuEeVeLgjsKPRL8WiBpLvlcEqSK5u50pwRFRlPvY8oBPUcv/
mybXzEj4hJK8Ty+L1HyAKGJi/RIy9rFB
-----END PUBLIC KEY-----`
)

type NerdClaims struct {
	*jwt.StandardClaims
	ProjectID string `json:"proj,omitempty"`
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

//decodeToken decodes a nerd token (JWT) given the public key to check if the signature is valid.
func decodeToken(nerdToken string, key *ecdsa.PublicKey) (*NerdClaims, error) {
	p := &jwt.Parser{
		SkipClaimsValidation: true,
	}
	token, err := p.ParseWithClaims(nerdToken, &NerdClaims{}, func(token *jwt.Token) (interface{}, error) {
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
