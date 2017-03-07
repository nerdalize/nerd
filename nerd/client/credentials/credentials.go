package credentials

import (
	"crypto/ecdsa"
	"sync"

	"github.com/pkg/errors"
)

//NerdAPI holds a reference to a nerdalize auth token. A credentials provider is needed to provide this value.
type NerdAPI struct {
	value        *NerdAPIValue
	PublicKey    *ecdsa.PublicKey
	provider     Provider
	forceRefresh bool
	m            sync.Mutex
}

type NerdAPIValue struct {
	NerdToken string
}

type Provider interface {
	IsExpired() bool
	Retrieve(*ecdsa.PublicKey) (*NerdAPIValue, error)
}

func NewNerdAPI(pub *ecdsa.PublicKey, provider Provider) *NerdAPI {
	return &NerdAPI{
		PublicKey: pub,
		provider:  provider,
		m:         sync.Mutex{},
	}
}

//Get the nerd token. This function checks with the provider whether the token is expired and if so retrieves a new token from the provider.
func (n *NerdAPI) Get() (*NerdAPIValue, error) {
	n.m.Lock()
	defer n.m.Unlock()

	if n.isExpired() {
		value, err := n.provider.Retrieve(n.PublicKey)
		if err != nil {
			return nil, errors.Wrap(err, "failed to retreive nerd api credentials")
		}
		n.value = value
		n.forceRefresh = false
	}

	return n.value, nil
}

func (n *NerdAPI) GetClaims() (*NerdClaims, error) {
	val, err := n.Get()
	if err != nil {
		return nil, errors.Wrap(err, "failed to retreive token")
	}
	claims, err := DecodeTokenWithKey(val.NerdToken, n.PublicKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode token")
	}
	return claims, nil
}

func (n *NerdAPI) isExpired() bool {
	return n.forceRefresh || n.provider.IsExpired()
}
