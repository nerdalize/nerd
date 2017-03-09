package aws

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/nerdalize/nerd/nerd/client"
	"github.com/pkg/errors"
)

// ProviderName is the name of the credentials provider.
const ProviderName = `NerdalizeProvider`

// Provider satisfies the credentials.Provider interface, and is a client to
// retrieve credentials from the nerdalize api.
type Provider struct {
	credentials.Expiry
	ExpiryWindow time.Duration
	Client       *client.NerdAPIClient
}

//NewNerdalizeCredentials creates a new credentials object with the NerdalizeProvider as provider.
func NewNerdalizeCredentials(c *client.NerdAPIClient) *credentials.Credentials {
	return credentials.NewCredentials(&Provider{
		Client: c,
	})
}

//IsExpired checks if the AWS sessions is expired.
func (p *Provider) IsExpired() bool {
	return p.Expiry.IsExpired()
}

// Retrieve will attempt to request the credentials from the nerdalize api.
// And error will be returned if the retrieval fails.
func (p *Provider) Retrieve() (credentials.Value, error) {
	sess, err := p.Client.CreateSession()
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
