package credentials

import (
	"crypto/ecdsa"
	"os"
	"time"

	v2client "github.com/nerdalize/nerd/nerd/client/v2"
	"github.com/pkg/errors"
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

//Retrieve retrieves the token from the env variable.
func (e *EnvProvider) Retrieve() (*v2client.Credentials, error) {
	token := os.Getenv(NerdTokenEnvVar)
	if token == "" {
		return nil, errors.Errorf("environment variable %v is not set", NerdTokenEnvVar)
	}
	claims, err := DecodeTokenWithKey(token, e.Pub)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode token '%v'", token)
	}
	e.AlwaysValid = claims.ExpiresAt == 0 // if unset
	e.SetExpiration(time.Unix(claims.ExpiresAt, 0))
	err = claims.Valid()
	if err != nil {
		return nil, errors.Wrapf(err, "nerd token '%v' is invalid", token)
	}
	return &v2client.Credentials{JWT: token}, nil
}
