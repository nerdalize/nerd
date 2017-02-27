package data

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/nerdalize/nerd/nerd/client"
	"github.com/nerdalize/nerd/nerd/client/credentials/provider"
	"github.com/pkg/errors"
)

// ProviderName is the name of the credentials provider.
const ProviderName = `NerdalizeProvider`
const JWTHomeLocation = `~/.nerd/jwt`
const JWTEnvName = `NERD_JWT`

// Provider satisfies the credentials.Provider interface, and is a client to
// retrieve credentials from an arbitrary endpoint.
type Provider struct {
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

	// TODO: these will be removed in the future
	AuthAPIURL string
	NerdAPIURL string
}

// NewCredentialsClient returns a Credentials wrapper for retrieving credentials
// from an arbitrary endpoint concurrently. The client will request the
func NewNerdalizeCredentials(nerdAPIURL string) *credentials.Credentials {
	return credentials.NewCredentials(&Provider{
		NerdAPIURL: nerdAPIURL,
	})
}

// IsExpired returns true if the credentials retrieved are expired, or not yet
// retrieved.
func (p *Provider) IsExpired() bool {
	return p.Expiry.IsExpired()
}

// Retrieve will attempt to request the credentials from the endpoint the Provider
// was configured for. And error will be returned if the retrieval fails.
func (p *Provider) Retrieve() (credentials.Value, error) {
	// TODO: Dont' only read from envdisk but also prompt for user input
	c := client.NewNerdAPIWithEndpoint(provider.NewEnvDiskCredentials(), p.NerdAPIURL)
	sess, err := c.CreateSession()
	if err != nil {
		return credentials.Value{ProviderName: ProviderName}, errors.Wrap(err, "failed to get AWS credentials")
	}

	p.SetExpiration(sess.AWSExpiration, p.ExpiryWindow)

	return credentials.Value{
		AccessKeyID:     sess.AWSAccessKeyID,
		SecretAccessKey: sess.AWSSecretAccessKey,
		SessionToken:    sess.AWSSessionToken,
		ProviderName:    ProviderName,
	}, nil
}
