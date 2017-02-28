package provider

import (
	"os"
	"time"

	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/pkg/errors"
)

//Env provides nerdalize credentials from the `credentials.NerdTokenEnvVar` environment variable.
type Env struct {
	*ProviderBasis
}

//NewEnvCredentials creates new nerdalize credentials with the env provider.
func NewEnvCredentials() *credentials.NerdAPI {
	return credentials.NewNerdAPI(NewEnv())
}

func NewEnv() *Env {
	return &Env{
		ProviderBasis: &ProviderBasis{
			ExpireWindow: DefaultExpireWindow,
		},
	}
}

func (e *Env) Retrieve() (*credentials.NerdAPIValue, error) {
	token := os.Getenv(credentials.NerdTokenEnvVar)
	if token == "" {
		return nil, errors.Errorf("environment variable %v is not set", credentials.NerdTokenEnvVar)
	}
	if e.TokenDecoder == nil {
		e.TokenDecoder = credentials.DecodeToken
	}
	claims, err := e.TokenDecoder(token)
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
