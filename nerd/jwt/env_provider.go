package jwt

import (
	"crypto/ecdsa"
	"os"
	"time"

	"github.com/pkg/errors"
)

const (
	//NerdTokenEnvVar is the environment variable used to set the JWT.
	NerdTokenEnvVar = "NERD_TOKEN"
)

//EnvProvider provides nerdalize credentials from the `credentials.NerdTokenEnvVar` environment variable.
type EnvProvider struct {
	*ProviderBasis
}

//NewEnvProvider creates a new EnvProvider provider.
func NewEnvProvider(pub *ecdsa.PublicKey) *EnvProvider {
	return &EnvProvider{
		ProviderBasis: &ProviderBasis{
			ExpireWindow: DefaultExpireWindow,
			Pub:          pub,
		},
	}
}

//Retrieve retrieves the jwt from the env variable.
func (e *EnvProvider) Retrieve() (string, error) {
	jwt := os.Getenv(NerdTokenEnvVar)
	if jwt == "" {
		return "", errors.Errorf("environment variable %v is not set", NerdTokenEnvVar)
	}
	claims, err := DecodeTokenWithKey(jwt, e.Pub)
	if err != nil {
		return "", errors.Wrapf(err, "failed to decode jwt '%v'", jwt)
	}
	e.AlwaysValid = claims.ExpiresAt == 0 // if unset
	e.SetExpiration(time.Unix(claims.ExpiresAt, 0))
	err = claims.Valid()
	if err != nil {
		return "", errors.Wrapf(err, "nerd jwt '%v' is invalid", jwt)
	}
	return jwt, nil
}
