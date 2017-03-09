package provider

import (
	"crypto/ecdsa"
	"os"
	"time"

	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/pkg/errors"
)

//Env provides nerdalize credentials from the `credentials.NerdTokenEnvVar` environment variable.
type Env struct {
	*Basis
}

//NewEnvCredentials creates new nerdalize credentials with the env provider.
func NewEnvCredentials(pub *ecdsa.PublicKey) *credentials.NerdAPI {
	return credentials.NewNerdAPI(pub, NewEnv())
}

//NewEnv creates a new Env provider.
func NewEnv() *Env {
	return &Env{
		Basis: &Basis{
			ExpireWindow: DefaultExpireWindow,
		},
	}
}

//Retrieve retrieves the token from the env variable.
func (e *Env) Retrieve(pub *ecdsa.PublicKey) (*credentials.NerdAPIValue, error) {
	token := os.Getenv(credentials.NerdTokenEnvVar)
	if token == "" {
		return nil, errors.Errorf("environment variable %v is not set", credentials.NerdTokenEnvVar)
	}
	claims, err := credentials.DecodeTokenWithKey(token, pub)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode token '%v'", token)
	}
	e.AlwaysValid = claims.ExpiresAt == 0 // if unset
	e.SetExpiration(time.Unix(claims.ExpiresAt, 0))
	err = claims.Valid()
	if err != nil {
		return nil, errors.Wrapf(err, "nerd token '%v' is invalid", token)
	}
	return &credentials.NerdAPIValue{NerdToken: token}, nil
}
