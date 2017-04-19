package provider

import (
	"crypto/ecdsa"
	"time"

	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/pkg/errors"
)

//Static provides nerdalize credentials from the a function argument.
type Static struct {
	*Basis
	JWT string
}

//NewStatic creates a new Static provider.
func NewStatic(jwt string) *Static {
	return &Static{
		JWT: jwt,
		Basis: &Basis{
			ExpireWindow: DefaultExpireWindow,
		},
	}
}

//Retrieve retrieves the token from the Static struct.
func (s *Static) Retrieve(pub *ecdsa.PublicKey) (*credentials.NerdAPIValue, error) {
	token := s.JWT
	if token == "" {
		return nil, errors.New("token not set")
	}
	claims, err := credentials.DecodeTokenWithKey(token, pub)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode token '%v'", token)
	}
	s.AlwaysValid = claims.ExpiresAt == 0 // if unset
	s.SetExpiration(time.Unix(claims.ExpiresAt, 0))
	err = claims.Valid()
	if err != nil {
		return nil, errors.Wrapf(err, "nerd token '%v' is invalid", token)
	}
	return &credentials.NerdAPIValue{NerdToken: token}, nil
}
