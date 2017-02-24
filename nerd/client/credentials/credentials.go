package credentials

import (
	"sync"

	"github.com/pkg/errors"
)

type NerdAPI struct {
	value        *NerdAPIValue
	provider     Provider
	forceRefresh bool
	m            sync.Mutex
}

type NerdAPIValue struct {
	NerdToken string
}

type Provider interface {
	IsExpired() bool
	Retrieve() (*NerdAPIValue, error)
}

func NewNerdAPI(provider Provider) *NerdAPI {
	return &NerdAPI{
		provider: provider,
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

func (n *NerdAPI) isExpired() bool {
	return n.forceRefresh || n.provider.IsExpired()
}
