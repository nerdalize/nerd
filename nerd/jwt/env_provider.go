package jwt

import (
	"crypto/ecdsa"
	"os"

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
	err := e.SetExpirationFromJWT(jwt)
	if err != nil {
		return "", errors.Wrap(err, "failed to set expiration")
	}
	return jwt, nil
}
