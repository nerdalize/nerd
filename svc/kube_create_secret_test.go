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

func TestCreateSecret(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Input    *svc.CreateSecretInput
		IsOutput func(tb testing.TB, out *svc.CreateSecretOutput)
		IsErr    func(error) bool
	}{
		{
			Name:    "when a zero value input is provided it should return a validation error",
			Timeout: time.Second * 5,
			Input:   nil,
			IsErr:   svc.IsValidationErr,
			IsOutput: func(t testing.TB, out *svc.CreateSecretOutput) {
				assert(t, out == nil, "output should be nil")
			},
		},
		{
			Name:    "when a valid input is provided it should return a secret with a unique name",
			Timeout: time.Second * 5,
			Input:   &svc.CreateSecretInput{Image: "smoketest", Project: "nerdalize", Registry: "quay.io", Username: "test", Password: "test"},
			IsErr:   nil,
			IsOutput: func(t testing.TB, out *svc.CreateSecretOutput) {
				assert(t, out != nil, "output should not be nil")
				assert(t, strings.Contains(out.Name, "s-"), "secret name should be generated and prefixed")
			},
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			di, clean := testDI(t)
			defer clean()

			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, c.Timeout)
			defer cancel()

			kube := svc.NewKube(di)
			out, err := kube.CreateSecret(ctx, c.Input)
			if c.IsErr != nil {
				assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			}

			if c.IsOutput != nil {
				c.IsOutput(t, out)
			}
		})
	}
}
