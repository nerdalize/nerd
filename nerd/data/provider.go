package data

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/nerdalize/nerd/nerd/client"
	"github.com/pkg/errors"
)

// ProviderName is the name of the credentials provider.
const ProviderName = `NerdalizeProvider`
const JWTHomeLocation = `~/.nerd/jwt`
const JWTEnvName = `NERD_JWT`

// Provider satisfies the credentials.Provider interface, and is a client to
// retrieve credentials from an arbitrary endpoint.
type Provider struct {
	staticCreds bool
	credentials.Expiry

	// ExpiryWindow will allow the credentials to trigger refreshing prior to
	// the credentials actually expiring. This is beneficial so race conditions
	// with expiring credentials do not cause request to fail unexpectedly
	// due to ExpiredTokenException exceptions.
	//
	// So a ExpiryWindow of 10s would cause calls to IsExpired() to return true
	// 10 seconds before the credentials are actually expired.
	//
	// If ExpiryWindow is 0 or less it will be ignored.
	ExpiryWindow time.Duration
}

// NewCredentialsClient returns a Credentials wrapper for retrieving credentials
// from an arbitrary endpoint concurrently. The client will request the
func NewNerdalizeCredentials() *credentials.Credentials {
	return credentials.NewCredentials(&Provider{staticCreds: false})
}

// IsExpired returns true if the credentials retrieved are expired, or not yet
// retrieved.
func (p *Provider) IsExpired() bool {
	if p.staticCreds {
		return false
	}
	return p.Expiry.IsExpired()
}

func getJWT() (string, error) {
	jwt := os.Getenv(JWTEnvName)
	if jwt != "" {
		return jwt, nil
	}
	if _, err := os.Stat(JWTHomeLocation); err == nil {
		dat, err := ioutil.ReadFile(JWTHomeLocation)
		if err != nil {
			return "", errors.Wrapf(err, "could not read '%v'", JWTHomeLocation)
		}
		return string(dat), nil
	}
	return "", errors.Errorf("the JWT could not be found in environment varibale '$%v' or in the file '%v'", JWTEnvName, JWTHomeLocation)
}

// Retrieve will attempt to request the credentials from the endpoint the Provider
// was configured for. And error will be returned if the retrieval fails.
func (p *Provider) Retrieve() (credentials.Value, error) {
	jwt, err := getJWT()
	if err != nil {
		return credentials.Value{ProviderName: ProviderName}, errors.Wrap(err, "failed to retrieve JWT")
	}
	c, err := client.NewNerdAPIFromJWT(jwt)
	if err != nil {
		return credentials.Value{ProviderName: ProviderName}, errors.Wrapf(err, "could not create client from JWT '%v'", jwt)
	}
	sess, err := c.CreateSession(jwt)
	if err != nil {
		return credentials.Value{ProviderName: ProviderName}, errors.Wrapf(err, "failed to get AWS credentials for JWT '%v'", jwt)
	}

	p.SetExpiration(sess.AWSExpiration, p.ExpiryWindow)

	return credentials.Value{
		AccessKeyID:     sess.AWSAccessKeyID,
		SecretAccessKey: sess.AWSSecretAccessKey,
		SessionToken:    sess.AWSSessionToken,
		ProviderName:    ProviderName,
	}, nil
}
