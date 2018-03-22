package svc_test

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/nerdalize/nerd/pkg/kubevisor"
	"github.com/nerdalize/nerd/svc"
)

func TestDeleteSecret(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Secrets  []*svc.CreateSecretInput
		Input    *svc.DeleteSecretInput
		Output   *svc.DeleteSecretOutput
		Listing  *svc.ListSecretsOutput
		IsOutput func(tb testing.TB, out *svc.DeleteSecretOutput, l *svc.ListSecretsOutput)
		IsErr    func(error) bool
	}{
		{
			Name:    "when no name is provided it should provide a validation error",
			Timeout: time.Second * 5,
			Input:   &svc.DeleteSecretInput{},
			Output:  &svc.DeleteSecretOutput{},
			IsErr:   svc.IsValidationErr,
		},
		{
			Name:    "when a non-existing secret is deleted it should return NotExists error",
			Timeout: time.Second * 5,
			Input:   &svc.DeleteSecretInput{Name: "foo"},
			Output:  &svc.DeleteSecretOutput{},
			IsErr:   kubevisor.IsNotExistsErr,
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
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

			out, err := kube.DeleteSecret(ctx, c.Input)
			if c.IsErr != nil {
				assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			}

			list, err := kube.ListSecrets(ctx, &svc.ListSecretsInput{})
			ok(t, err)

			if c.IsOutput != nil {
				c.IsOutput(t, out, list)
			}
		})
	}
}

func TestDeleteSpecificSecret(t *testing.T) {
	image := "quay.io/nerdalize/smoketest"
	timeout := time.Minute

	if testing.Short() {
		t.Skipf("skipping long test with contex timeout: %s", timeout)
	}

	di, clean := testDI(t)
	defer clean()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	kube := svc.NewKube(di)
	secret, err := kube.CreateSecret(ctx, &svc.CreateSecretInput{Image: image, Username: "test", Password: "test"})
	ok(t, err)

	out, err := kube.DeleteSecret(ctx, &svc.DeleteSecretInput{Name: secret.Name})
	ok(t, err)
	assert(t, out != nil, "expected to find a DeleteSecretOutput")
}
