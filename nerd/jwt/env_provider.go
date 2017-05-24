package jwt

import (
	"crypto/ecdsa"
	"os"

	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

const (
	//NerdTokenEnvVar is the environment variable used to set the JWT
	NerdTokenEnvVar = "NERD_JWT"
	//NerdSecretEnvVar is the environment variable used for the JWT refresh secret
	NerdSecretEnvVar = "NERD_JWT_REFRESH_TOKEN"
)

//EnvProvider provides nerdalize credentials from the `credentials.NerdTokenEnvVar` environment variable.
type EnvProvider struct {
	*ProviderBasis
	Client  v1auth.TokenClientInterface
	Session conf.SessionInterface
}

//NewEnvProvider creates a new EnvProvider provider.
func NewEnvProvider(pub *ecdsa.PublicKey, session conf.SessionInterface, client v1auth.TokenClientInterface) *EnvProvider {
	return &EnvProvider{
		ProviderBasis: &ProviderBasis{
			ExpireWindow: DefaultExpireWindow,
			Pub:          pub,
		},
		Session: session,
		Client:  client,
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

	jwtSecret := os.Getenv(NerdSecretEnvVar)
	if jwtSecret != "" && e.IsExpired() {
		jwt, err = e.refresh(jwt, jwtSecret)
		if err != nil {
			return "", errors.Wrap(err, "failed to refresh")
		}
	}
	err = isValid(jwt, e.Pub)
	if err != nil {
		return "", err
	}
	return jwt, nil
}

func (e *EnvProvider) refresh(jwt, secret string) (string, error) {
	ss, err := e.Session.Read()
	if err != nil {
		return "", errors.Wrap(err, "failed to read config")
	}
	out, err := e.Client.RefreshJWT(ss.Project.Name, jwt, secret)
	if err != nil {
		return "", errors.Wrap(err, "failed to refresh token")
	}
	err = e.SetExpirationFromJWT(out.Token)
	if err != nil {
		return "", errors.Wrap(err, "failed to set expiration")
	}
	return out.Token, nil
}
