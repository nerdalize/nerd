package credentials

import (
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
)

type EnvProvider struct {
	retrieved bool
}

func NewEnvProvider() *EnvProvider {
	return &EnvProvider{}
}

func (e *EnvProvider) IsExpired() bool {
	// TODO: check for expiration in JWT
	return !e.retrieved
}

func (e *EnvProvider) Retrieve() (*NerdAPIValue, error) {
	// TODO: check for expiration in JWT
	e.retrieved = false

	// TODO: magic name
	token := os.Getenv("NERD_TOKEN")
	if token != "" {
		e.retrieved = true
		return &NerdAPIValue{NerdToken: token}, nil
	}
	t, err := ioutil.ReadFile(JWTHomeLocation)
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "could not read nerd token at '%v'", JWTHomeLocation)
	}
	e.retrieved = true
	return &NerdAPIValue{NerdToken: string(t)}, nil
}
