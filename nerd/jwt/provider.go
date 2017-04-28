package jwt

import (
	"crypto/ecdsa"
	"time"

	"github.com/pkg/errors"
)

//DefaultExpireWindow is the default amount of seconds that a nerd token is assumed to be expired, before it's actually expired.
//This will prevent the server from declining the token because it was just expired.
const DefaultExpireWindow = 20

//ProviderBasis is the basis for every provider.
type ProviderBasis struct {
	expiration  time.Time
	CurrentTime func() time.Time
	AlwaysValid bool

	ExpireWindow time.Duration

	Pub *ecdsa.PublicKey
}

//IsExpired checks if the current token is expired.
func (b *ProviderBasis) IsExpired() bool {
	if b.CurrentTime == nil {
		b.CurrentTime = time.Now
	}
	return !b.AlwaysValid && !b.CurrentTime().Before(b.expiration)
}

//SetExpiration sets the expiration field and takes the ExpireWindow into account.
func (b *ProviderBasis) SetExpiration(expiration time.Time) {
	b.expiration = expiration
	if b.ExpireWindow > 0 {
		b.expiration = b.expiration.Add(-b.ExpireWindow)
	}
}

//SetExpirationFromJWT decodes the JWT and sets the provider expiration based on the JWT expiration field.
func (b *ProviderBasis) SetExpirationFromJWT(jwt string) error {
	claims, err := DecodeTokenWithKey(jwt, b.Pub)
	if err != nil {
		return errors.Wrapf(err, "failed to decode jwt '%v'", jwt)
	}
	b.AlwaysValid = claims.ExpiresAt == 0 // if unset
	b.SetExpiration(time.Unix(claims.ExpiresAt, 0))
	err = claims.Valid()
	if err != nil {
		return errors.Wrapf(err, "nerd jwt '%v' is invalid", jwt)
	}
	return nil
}
