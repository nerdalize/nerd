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

func TestGetSecret(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Secrets  []*svc.CreateSecretInput
		Input    *svc.GetSecretInput
		IsOutput func(tb testing.TB, out *svc.GetSecretOutput) bool
		IsErr    func(error) bool
	}{
		{
			Name:    "when a zero value input is provided it should return a validation error",
			Timeout: time.Second * 5,
			Secrets: nil,
			Input:   nil,
			IsErr:   svc.IsValidationErr,
			IsOutput: func(t testing.TB, out *svc.GetSecretOutput) bool {
				return true
			},
		},
		{
			Name:    "when secret doesnt exist it should return an error",
			Timeout: time.Second * 5,
			Input:   &svc.GetSecretInput{Name: "my-secret"},
			IsErr:   kubevisor.IsNotExistsErr,
			IsOutput: func(t testing.TB, out *svc.GetSecretOutput) bool {
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
			for _, secrets := range c.Secrets {
				_, err := kube.CreateSecret(ctx, secrets)
				ok(t, err)
			}

			out, err := kube.GetSecret(ctx, c.Input)
			if c.IsErr != nil { //if c.IsErr is nil we dont care about errors
				assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			}

			if c.IsOutput != nil {
				c.IsOutput(t, out)
			}
		})
	}
}

func TestGetSpecificSecret(t *testing.T) {
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

	out, err := kube.GetSecret(ctx, &svc.GetSecretInput{Name: secret.Name})
	ok(t, err)
	assert(t, out != nil, "expected to find a secret")
	assert(t, out.Image == image, "expected to find the base image in the secret")

}
