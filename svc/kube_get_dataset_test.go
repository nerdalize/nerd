package svc_test

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/nerdalize/nerd/pkg/kubevisor"
	"github.com/nerdalize/nerd/pkg/transfer/archiver"
	"github.com/nerdalize/nerd/pkg/transfer/store"
	"github.com/nerdalize/nerd/svc"
)

func TestGetDataset(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Datasets []*svc.CreateDatasetInput
		Input    *svc.GetDatasetInput
		IsOutput func(tb testing.TB, out *svc.GetDatasetOutput) bool
		IsErr    func(error) bool
	}{
		{
			Name:     "when a zero value input is provided it should return a validation error",
			Timeout:  time.Second * 5,
			Datasets: nil,
			Input:    nil,
			IsErr:    svc.IsValidationErr,
			IsOutput: func(t testing.TB, out *svc.GetDatasetOutput) bool {
				return true
			},
		},
		{
			Name:    "when dataset doesnt exist it should return an error",
			Timeout: time.Second * 5,
			Input:   &svc.GetDatasetInput{Name: "my-dataset"},
			IsErr:   kubevisor.IsNotExistsErr,
			IsOutput: func(t testing.TB, out *svc.GetDatasetOutput) bool {
				return true
			},
		},
		{
			Name:     "when one dataset has been uploaded it should be available for download",
			Timeout:  time.Minute,
			Datasets: []*svc.CreateDatasetInput{{Name: "my-dataset", StoreOptions: transferstore.StoreOptions{Type: transferstore.StoreTypeS3}, ArchiverOptions: transferarchiver.ArchiverOptions{Type: transferarchiver.ArchiverTypeTar}}},
			Input:    &svc.GetDatasetInput{Name: "my-datasets"},
			IsErr:    nil,
			IsOutput: func(t testing.TB, out *svc.GetDatasetOutput) bool {
				if out == nil {
					return false
				}
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
			for _, datasets := range c.Datasets {
				_, err := kube.CreateDataset(ctx, datasets)
				ok(t, err)
			}

			out, err := kube.GetDataset(ctx, c.Input)
			if c.IsErr != nil { //if c.IsErr is nil we dont care about errors
				assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			}

			if c.IsOutput != nil {
				c.IsOutput(t, out)
			}
		})
	}
}
