package provider

import (
	"io/ioutil"
	"os"

	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/pkg/errors"
)

type EnvDiskProvider struct {
	retrieved bool
}

func NewEnvDiskProvider() *EnvDiskProvider {
	return &EnvDiskProvider{}
}

func (e *EnvDiskProvider) IsExpired() bool {
	// TODO: check for expiration in JWT
	return !e.retrieved
}

func (e *EnvDiskProvider) Retrieve() (*credentials.NerdAPIValue, error) {
	// TODO: check for expiration in JWT
	e.retrieved = false

	token := os.Getenv(credentials.NerdTokenEnvVar)
	if token != "" {
		e.retrieved = true
		return &credentials.NerdAPIValue{NerdToken: token}, nil
	}
	filename, err := credentials.TokenFilename()
	if err != nil {
		return nil, err
	}
	t, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "both env variable %v and disk location %v are empty", credentials.NerdTokenEnvVar, filename)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "could not read nerd token at '%v'", filename)
	}
	e.retrieved = true
	return &credentials.NerdAPIValue{NerdToken: string(t)}, nil
}
