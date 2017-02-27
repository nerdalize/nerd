package provider

import (
	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/pkg/errors"
)

type ChainProvider struct {
	Providers []credentials.Provider
	curr      credentials.Provider
}

// NewChainCredentials returns a pointer to a new Credentials object
// wrapping a chain of providers.
func NewChainCredentials(providers ...credentials.Provider) *credentials.NerdAPI {
	return credentials.NewNerdAPI(&ChainProvider{
		Providers: providers,
	})
}

// Retrieve returns the credentials value or error if no provider returned
// without error.
//
// If a provider is found it will be cached and any calls to IsExpired()
// will return the expired state of the cached provider.
func (c *ChainProvider) Retrieve() (*credentials.NerdAPIValue, error) {
	var errs []error
	for _, p := range c.Providers {
		creds, err := p.Retrieve()
		if err == nil {
			c.curr = p
			return creds, nil
		}
		errs = append(errs, err)
	}
	c.curr = nil

	return nil, errors.Errorf("could not retreive token from any provider: %v", errs)
}

// IsExpired will returned the expired state of the currently cached provider
// if there is one.  If there is no current provider, true will be returned.
func (c *ChainProvider) IsExpired() bool {
	if c.curr != nil {
		return c.curr.IsExpired()
	}

	return true
}
