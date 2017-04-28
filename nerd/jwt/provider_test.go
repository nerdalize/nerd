package jwt

import (
	"testing"
	"time"
)

func TestBasisProvider(t *testing.T) {
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
