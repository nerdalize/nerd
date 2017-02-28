package provider

import (
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/pkg/errors"
)

const DefaultExpireWindow = 20

//ProviderBasis is the basis for every provider.
type ProviderBasis struct {
	expiration  time.Time
	CurrentTime func() time.Time
	AlwaysValid bool

	ExpireWindow time.Duration
	TokenDecoder func(string) (*credentials.NerdClaims, error)
}

//IsExpired checks if the current token is expired.
func (b *ProviderBasis) IsExpired() bool {
	if b.CurrentTime == nil {
		b.CurrentTime = time.Now
	}
	return b.AlwaysValid || b.expiration.Before(b.CurrentTime())
}

//SetExpiration sets the expiration field and takes the ExpireWindow into account.
func (b *ProviderBasis) SetExpiration(expiration time.Time) {
	b.expiration = expiration
	if b.ExpireWindow > 0 {
		b.expiration = b.expiration.Add(-b.ExpireWindow)
	}
}

func TokenFilename() (string, error) {
	f, err := homedir.Expand("~/.nerd/token")
	if err != nil {
		return "", errors.Wrap(err, "failed to retreive homedir path")
	}
	return f, nil
}
