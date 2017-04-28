package jwt

import (
	"crypto/ecdsa"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func TestExpiration(t *testing.T) {
	cases := []struct {
		alwaysValid bool
		window      time.Duration
		expireAt    time.Duration
		expected    bool
	}{
		{false, 0, 0, true},
		{false, 0, +1 * time.Second, false},
		{false, 0, +10 * time.Second, false},
		{false, 0, -1 * time.Second, true},
		{false, 0, -10 * time.Second, true},
		{false, 10 * time.Second, 0, true},
		{false, 10 * time.Second, +1 * time.Second, true},
		{false, 10 * time.Second, +10 * time.Second, true},
		{false, 10 * time.Second, +11 * time.Second, false},
		{false, 10 * time.Second, -1 * time.Second, true},
		{false, 10 * time.Second, -10 * time.Second, true},

		{true, 0, 0, false},
		{true, 0, +1 * time.Second, false},
		{true, 0, +10 * time.Second, false},
		{true, 0, -1 * time.Second, false},
		{true, 0, -10 * time.Second, false},
		{true, 10 * time.Second, 0, false},
		{true, 10 * time.Second, +1 * time.Second, false},
		{true, 10 * time.Second, +10 * time.Second, false},
		{true, 10 * time.Second, +11 * time.Second, false},
		{true, 10 * time.Second, -1 * time.Second, false},
		{true, 10 * time.Second, -10 * time.Second, false},
	}
	for _, c := range cases {
		now := time.Now()
		basis := &ProviderBasis{
			CurrentTime: func() time.Time {
				return now
			},
			AlwaysValid:  c.alwaysValid,
			ExpireWindow: c.window,
		}
		basis.SetExpiration(now.Add(c.expireAt))
		if basis.IsExpired() != c.expected {
			t.Errorf("Expected %v but got %v for case %v", c.expected, basis.IsExpired(), c)
		}
	}
}

func TestExpirationFromJWT(t *testing.T) {
	key := testkey(t)
	exp := time.Unix(time.Now().Add(time.Minute).Unix(), 0) // we need a little hack to make sure we round to seconds
	token := jwt.NewWithClaims(jwt.SigningMethodES384, &jwt.StandardClaims{
		ExpiresAt: exp.Unix(),
	})
	ss, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign test token: %v", err)
	}
	pub, _ := key.Public().(*ecdsa.PublicKey)
	basis := &ProviderBasis{
		Pub: pub,
	}
	err = basis.SetExpirationFromJWT(ss)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if basis.expiration != exp {
		t.Errorf("expected expiration %v but got %v", exp, basis.expiration)
	}
}
