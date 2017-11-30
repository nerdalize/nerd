package svc_test

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/nerdalize/nerd/svc"
)

func TestRunJob(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Input    *svc.RunJobInput
		Output   *svc.RunJobOutput
		IsOutput func(tb testing.TB, out *svc.RunJobOutput)
		IsErr    func(error) bool
	}{
		{
			Name:    "when a zero value input is provided it should return a no input error",
			Timeout: time.Second * 5,
			Input:   nil,
			IsErr:   svc.IsNoInputErr,
			IsOutput: func(tb testing.TB, out *svc.RunJobOutput) {
				assert(tb, out == nil, "output should be nil")
			},
		},
		{
			Name:    "when input is provided that is invalid it should return a validation error",
			Timeout: time.Second * 5,
			Input:   &svc.RunJobInput{},
			IsErr:   svc.IsValidationErr,
			IsOutput: func(tb testing.TB, out *svc.RunJobOutput) {
				assert(tb, out == nil, "output should be nil")
			},
		},
		{
			Name:    "when a job is started with just an image it should generate a name and return it",
			Timeout: time.Second * 5,
			Input:   &svc.RunJobInput{Image: "hello-world"},
			IsErr:   isNilErr,
			IsOutput: func(tb testing.TB, out *svc.RunJobOutput) {
				assert(tb, out != nil, "output should not be nil")
				assert(tb, regexp.MustCompile(`^j-.+$`).MatchString(out.Name), "name should have a prefix but not be empty after the prefix")
			},
		},
		{
			Name:    "when a job is started with a very short deadline it should return a specific error",
			Timeout: time.Millisecond,
			Input:   &svc.RunJobInput{Image: "hello-world"},
			IsErr:   svc.IsDeadlineErr,
		},
		{
			Name:    "when a job is started with an invalid name it should return a invalid name error",
			Timeout: time.Second * 5,
			Input:   &svc.RunJobInput{Image: "hello-world", Name: "my-name-"},
			IsErr:   svc.IsInvalidNameErr,
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			di := testDI(t)
			ns, clean := testNamespace(t, di.Kube())
			defer clean()

			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, c.Timeout)
			defer cancel()

			kube := svc.NewKube(di, ns)
			out, err := kube.RunJob(ctx, c.Input)
			if c.IsErr != nil {
				assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			}

			if c.IsOutput != nil {
				c.IsOutput(t, out)
			}
		})
	}
}

func TestRunJobWithoutTheNamespace(t *testing.T) {
	di := testDI(t)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	kube := svc.NewKube(di, "non-existing")
	_, err := kube.RunJob(ctx, &svc.RunJobInput{Image: "hello-world", Name: "my-job"})
	assert(t, svc.IsNamespaceNotExistsErr(err), "expected error to be namespace doesn't exist")
}

func TestRunJobWithNameThatAlreadyExists(t *testing.T) {
	di := testDI(t)
	ns, clean := testNamespace(t, di.Kube())
	defer clean()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	kube := svc.NewKube(di, ns)
	out, err := kube.RunJob(ctx, &svc.RunJobInput{Image: "hello-world", Name: "my-job"})
	ok(t, err)

	_, err = kube.RunJob(ctx, &svc.RunJobInput{Image: "hello-world", Name: out.Name})
	assert(t, svc.IsAlreadyExistsErr(err), "expected error to be already exists")
}
