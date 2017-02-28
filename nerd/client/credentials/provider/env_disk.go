package provider

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/pkg/errors"
)

type EnvDisk struct {
	// The date/time when to expire on
	expiration  time.Time
	CurrentTime func() time.Time
	AlwaysValid bool

	DiskLocation string

	ExpireWindow time.Duration
	TokenDecoder func(string) (*credentials.NerdClaims, error)
}

// NewChainCredentials returns a pointer to a new Credentials object
// wrapping a chain of providers.
func NewEnvDiskCredentials() *credentials.NerdAPI {
	return credentials.NewNerdAPI(NewEnvDisk())
}

func NewEnvDisk() *EnvDisk {
	return &EnvDisk{
		AlwaysValid: false,
	}
}

func (e *EnvDisk) IsExpired() bool {
	if e.CurrentTime == nil {
		e.CurrentTime = time.Now
	}
	return e.AlwaysValid || e.expiration.Before(e.CurrentTime())
}

func (e *EnvDisk) SetExpiration(expiration time.Time) {
	e.expiration = expiration
	if e.ExpireWindow > 0 {
		e.expiration = e.expiration.Add(-e.ExpireWindow)
	}
}

func (e *EnvDisk) Retrieve() (*credentials.NerdAPIValue, error) {
	if e.DiskLocation == "" {
		filename, err := credentials.TokenFilename()
		if err != nil {
			return nil, err
		}
		e.DiskLocation = filename
	}
	t, err := ioutil.ReadFile(e.DiskLocation)
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "could not read nerd token at '%v'", e.DiskLocation)
	}
	token := ""
	if !os.IsNotExist(err) {
		token = string(t)
	}
	envtoken := os.Getenv(credentials.NerdTokenEnvVar)
	if envtoken != "" {
		token = envtoken
	}
	if token == "" {
		return nil, errors.Errorf("both env variable %v and disk location %v are empty", credentials.NerdTokenEnvVar, e.DiskLocation)
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
