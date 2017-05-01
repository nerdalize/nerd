package aws

import (
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v1batch "github.com/nerdalize/nerd/nerd/client/batch/v1"
	"github.com/pkg/errors"
)

// ProviderName is the name of the credentials provider.
const ProviderName = `NerdalizeProvider`

//DefaultExpireWindow is the default amount of seconds that the credentials are assumed to be expired, before they are actually expired.
//This will prevent the server from rejecting the credentials because they were just expired.
const DefaultExpireWindow = 20

// Provider satisfies the credentials.Provider interface, and is a client to
// retrieve credentials from the nerdalize api.
type Provider struct {
	credentials.Expiry
	ExpiryWindow time.Duration
	Client       v1batch.ClientTokenInterface
	NlzProjectID string
}

//NewNerdalizeCredentials creates a new credentials object with the NerdalizeProvider as provider.
func NewNerdalizeCredentials(c v1batch.ClientTokenInterface, nlzProjectID string) *credentials.Credentials {
	return credentials.NewCredentials(&Provider{
		Client:       c,
		NlzProjectID: nlzProjectID,
		ExpiryWindow: DefaultExpireWindow,
	})
}

//IsExpired checks if the AWS sessions is expired.
func (p *Provider) IsExpired() bool {
	return p.Expiry.IsExpired()
}

// Retrieve will attempt to request the credentials from the nerdalize api.
// And error will be returned if the retrieval fails.
func (p *Provider) Retrieve() (credentials.Value, error) {
	token, err := p.Client.CreateToken(p.NlzProjectID)
	if err != nil {
		return credentials.Value{ProviderName: ProviderName}, errors.Wrap(err, "failed to get AWS credentials")
	}

	p.SetExpiration(token.AWSExpiration, p.ExpiryWindow)

	return credentials.Value{
		AccessKeyID:     token.AWSAccessKeyID,
		SecretAccessKey: token.AWSSecretAccessKey,
		SessionToken:    token.AWSSessionToken,
		ProviderName:    ProviderName,
	}, nil
}
