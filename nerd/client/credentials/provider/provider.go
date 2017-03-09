package provider

import (
	"time"
)

//DefaultExpireWindow is the default amount of seconds that a nerd token is assumed to be expired, before it's actually expired.
//This will prevent the server from declining the token because it was just expired.
const DefaultExpireWindow = 20

//Basis is the basis for every provider.
type Basis struct {
	expiration  time.Time
	CurrentTime func() time.Time
	AlwaysValid bool

	ExpireWindow time.Duration
}

//IsExpired checks if the current token is expired.
func (b *Basis) IsExpired() bool {
	if b.CurrentTime == nil {
		b.CurrentTime = time.Now
	}
	return b.AlwaysValid || b.expiration.Before(b.CurrentTime())
}

//SetExpiration sets the expiration field and takes the ExpireWindow into account.
func (b *Basis) SetExpiration(expiration time.Time) {
	b.expiration = expiration
	if b.ExpireWindow > 0 {
		b.expiration = b.expiration.Add(-b.ExpireWindow)
	}
}
