package svc_test

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"testing"

	"github.com/nerdalize/nerd/svc"
)

func TestRunJob(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Ctx      context.Context
		Input    *svc.RunJobInput
		Output   *svc.RunJobOutput
		IsOutput func(tb testing.TB, out *svc.RunJobOutput)
		IsErr    func(error) bool
	}{
		{
			Name:  "when a zero value input is provided it should return a no input error",
			Ctx:   context.Background(),
			Input: nil,
			IsErr: svc.IsNoInputErr,
			IsOutput: func(tb testing.TB, out *svc.RunJobOutput) {
				assert(tb, out == nil, "output should be nil")
			},
		},
		{
			Name:  "when input is provided that is invalid it should return a validation error",
			Ctx:   context.Background(),
			Input: &svc.RunJobInput{},
			IsErr: svc.IsValidationErr,
			IsOutput: func(tb testing.TB, out *svc.RunJobOutput) {
				assert(tb, out == nil, "output should be nil")
			},
		},
		{
			Name:  "when a job is started with just an image it should generate a name and return it",
			Ctx:   context.Background(),
			Input: &svc.RunJobInput{Image: "hello-world"},
			IsErr: isNilErr,
			IsOutput: func(tb testing.TB, out *svc.RunJobOutput) {
				assert(tb, out != nil, "output should not be nil")
				assert(tb, regexp.MustCompile(`^j-.+$`).MatchString(out.Name), "name should have a prefix but not be empty after the prefix")
			},
		},
		//@TODO test the usecase of the usecase that doesn't exist
	} {
		t.Run(c.Name, func(t *testing.T) {
			di := testDI(t)
			ns, clean := testNamespace(t, di.Kube())
			defer clean()

			kube := svc.NewKube(di, ns)
			out, err := kube.RunJob(c.Ctx, c.Input)
			if c.IsErr != nil {
				assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			}

			if c.IsOutput != nil {
				c.IsOutput(t, out)
			}
		})
	}

	// t.Run("when no namespace is available, it should return a specific error", func(t *testing.T) {
	//
	// })
	//
	// t.Run("name", func(t *testing.T) {
	//
	// })

	//@TODO test the case in which no namespace is available
}
