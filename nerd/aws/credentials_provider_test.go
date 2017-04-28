package aws

import (
	"testing"
	"time"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

type fakeClient struct {
	exp time.Time
}

func (f *fakeClient) CreateToken(projectID string) (output *v1payload.CreateTokenOutput, err error) {
	return &v1payload.CreateTokenOutput{
		AWSExpiration: f.exp,
	}, nil
}

func TestExpired(t *testing.T) {
	cases := map[string]struct {
		expireAt time.Time
		expected bool
	}{
		"expired": {
			expireAt: time.Now().Add(-time.Minute),
			expected: true,
		},
		"notExpired": {
			expireAt: time.Now().Add(time.Minute),
			expected: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			cl := &fakeClient{tc.expireAt}
			prov := &Provider{
				Client: cl,
			}
			_, err := prov.Retrieve()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expired := prov.IsExpired()
			if expired != tc.expected {
				t.Errorf("expected %v but got %v", tc.expected, expired)
			}
		})
	}

}
