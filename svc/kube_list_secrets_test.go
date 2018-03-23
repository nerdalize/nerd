package svc_test

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/nerdalize/nerd/svc"
)

func TestListSecrets(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Secrets  []*svc.CreateSecretInput
		Input    *svc.ListSecretsInput
		IsOutput func(tb testing.TB, out *svc.ListSecretsOutput) bool
		IsErr    func(error) bool
	}{
		{
			Name:    "when a zero value input is provided it should return a validation error",
			Timeout: time.Second * 5,
			Secrets: nil,
			Input:   nil,
			IsErr:   svc.IsValidationErr,
			IsOutput: func(t testing.TB, out *svc.ListSecretsOutput) bool {
				return true
			},
		},
		{
			Name:    "when no secrets have been created the output should be empty",
			Timeout: time.Second * 5,
			Input:   &svc.ListSecretsInput{},
			IsErr:   isNilErr,
			IsOutput: func(t testing.TB, out *svc.ListSecretsOutput) bool {
				assert(t, len(out.Items) == 0, "expected zero secrets to be listed")
				return true
			},
		},
		{
			Name:    "when one correct secret was created it should be listed",
			Timeout: time.Minute,
			Secrets: []*svc.CreateSecretInput{{Image: "smoketest", Project: "nerdalize", Registry: "quay.io", Username: "test", Password: "test"}},
			Input:   &svc.ListSecretsInput{},
			IsErr:   isNilErr,
			IsOutput: func(t testing.TB, out *svc.ListSecretsOutput) bool {
				assert(t, len(out.Items) == 1, "expected one secret to be listed")
				assert(t, !out.Items[0].Details.CreatedAt.IsZero(), "created at time should not be zero")
				assert(t, out.Items[0].Details.Image == "quay.io/nerdalize/smoketest", "expected to find complete image name")
				assert(t, strings.HasPrefix(out.Items[0].Name, "s-"), "expected secret name to be prefixed has expected")
				return true
			},
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			if c.Timeout > time.Second*5 && testing.Short() {
				t.Skipf("skipping long test with contex timeout: %s", c.Timeout)
			}

			di, clean := testDI(t)
			defer clean()

			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, c.Timeout)
			defer cancel()

			kube := svc.NewKube(di)
			for _, secret := range c.Secrets {
				_, err := kube.CreateSecret(ctx, secret)
				ok(t, err)
			}

			out, err := kube.ListSecrets(ctx, c.Input)
			if c.IsErr != nil {
				assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			}

			if c.IsOutput != nil {
				c.IsOutput(t, out)
			}
		})
	}
}
