package svc_test

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/nerdalize/nerd/pkg/kubevisor"
	"github.com/nerdalize/nerd/svc"
)

func TestUploadDataset(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Input    *svc.UploadDatasetInput
		Output   *svc.UploadDatasetOutput
		IsOutput func(tb testing.TB, out *svc.UploadDatasetOutput)
		IsErr    func(error) bool
	}{
		{
			Name:    "when a zero value input is provided it should return a validation error",
			Timeout: time.Second * 5,
			Input:   nil,
			IsErr:   svc.IsValidationErr,
			IsOutput: func(t testing.TB, out *svc.UploadDatasetOutput) {
				assert(t, out == nil, "output should be nil")
			},
		},
		{
			Name:    "when input is provided that is invalid it should return a validation error",
			Timeout: time.Second * 5,
			Input:   &svc.UploadDatasetInput{},
			IsErr:   svc.IsValidationErr,
			IsOutput: func(t testing.TB, out *svc.UploadDatasetOutput) {
				assert(t, out == nil, "output should be nil")
			},
		},
		{
			Name:    "when a dataset is uploaded with just an input dir it should generate a name and return it",
			Timeout: time.Second * 5,
			Input:   &svc.UploadDatasetInput{Dir: "/tmp"},
			IsErr:   isNilErr,
			IsOutput: func(t testing.TB, out *svc.UploadDatasetOutput) {
				assert(t, out != nil, "output should not be nil")
				assert(t, regexp.MustCompile(`^d-.+$`).MatchString(out.Name), "name should have a prefix but not be empty after the prefix")
			},
		},
		{
			Name:    "when a dataset is uploaded with an invalid name it should return a invalid name error",
			Timeout: time.Second * 5,
			Input:   &svc.UploadDatasetInput{Name: "my-name-", Dir: "/tmp"},
			IsErr:   kubevisor.IsInvalidNameErr,
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			di, clean := testDI(t)
			defer clean()

			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, c.Timeout)
			defer cancel()

			kube := svc.NewKube(di)
			out, err := kube.UploadDataset(ctx, c.Input)
			if c.IsErr != nil {
				assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			}

			if c.IsOutput != nil {
				c.IsOutput(t, out)
			}
		})
	}
}

func TestUploadDatasetWithNameThatAlreadyExists(t *testing.T) {
	di, clean := testDI(t)
	defer clean()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	kube := svc.NewKube(di)
	out, err := kube.UploadDataset(ctx, &svc.UploadDatasetInput{Name: "my-dataset", Dir: "/tmp"})
	ok(t, err)

	_, err = kube.UploadDataset(ctx, &svc.UploadDatasetInput{Name: out.Name, Dir: "/tmp"})
	assert(t, kubevisor.IsAlreadyExistsErr(err), "expected error to be already exists")
}
