package svc_test

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/nerdalize/nerd/pkg/transfer/archiver"
	"github.com/nerdalize/nerd/pkg/transfer/store"
	"github.com/nerdalize/nerd/svc"
)

func TestListDatasets(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Datasets []*svc.CreateDatasetInput
		Input    *svc.ListDatasetsInput
		IsOutput func(tb testing.TB, out *svc.ListDatasetsOutput) bool
		IsErr    func(error) bool
	}{
		{
			Name:     "when a zero value input is provided it should return a validation error",
			Timeout:  time.Second * 5,
			Datasets: nil,
			Input:    nil,
			IsErr:    svc.IsValidationErr,
			IsOutput: func(t testing.TB, out *svc.ListDatasetsOutput) bool {
				return true
			},
		},
		{
			Name:    "when no datasets have been uploaded the output should be empty",
			Timeout: time.Second * 5,
			Input:   &svc.ListDatasetsInput{},
			IsErr:   isNilErr,
			IsOutput: func(t testing.TB, out *svc.ListDatasetsOutput) bool {
				assert(t, len(out.Items) == 0, "expected zero datasets to be listed")
				return true
			},
		},
		{
			Name:     "when one correct dataset was uploaded it should be listed",
			Timeout:  time.Minute,
			Datasets: []*svc.CreateDatasetInput{{Name: "my-dataset", StoreOptions: transferstore.StoreOptions{Type: transferstore.StoreTypeS3}, ArchiverOptions: transferarchiver.ArchiverOptions{Type: transferarchiver.ArchiverTypeTar}}},
			Input:    &svc.ListDatasetsInput{},
			IsErr:    isNilErr,
			IsOutput: func(t testing.TB, out *svc.ListDatasetsOutput) bool {
				assert(t, len(out.Items) == 1, "expected one dataset to be listed")
				assert(t, !out.Items[0].Details.CreatedAt.IsZero(), "created at time should not be zero")

				assert(t, out.Items[0].Name == "my-dataset", "expected dataset name to be equal to what was provided")
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
			for _, dataset := range c.Datasets {
				_, err := kube.CreateDataset(ctx, dataset)
				ok(t, err)
			}

			out, err := kube.ListDatasets(ctx, c.Input)
			if c.IsErr != nil {
				assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			}

			if c.IsOutput != nil {
				c.IsOutput(t, out)
			}
		})
	}
}
