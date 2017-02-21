package credentials

import (
	"sync"

	"github.com/pkg/errors"
)

type NerdAPI struct {
	value        *NerdAPIValue
	provider     *NerdAPIProvider
	forceRefresh bool
	m            sync.Mutex
}

type NerdAPIValue struct {
	NerdToken string
}

func NewNerdAPI() *NerdAPI {
	return &NerdAPI{
		provider: NewNerdAPIProvider(),
		m:        sync.Mutex{},
	}
}

func (n *NerdAPI) Get() (*NerdAPIValue, error) {
	n.m.Lock()
	defer n.m.Unlock()

	if n.isExpired() {
		value, err := n.provider.Retrieve()
		if err != nil {
			return nil, errors.Wrap(err, "failed to retreive nerd api credentials")
		}
		n.value = value
		n.forceRefresh = false
	}

	return n.value, nil
}

// isExpired helper method wrapping the definition of expired credentials.
func (n *NerdAPI) isExpired() bool {
	return n.forceRefresh || n.provider.IsExpired()
}
