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

func TestDownloadDataset(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Datasets []*svc.UploadDatasetInput
		Input    *svc.DownloadDatasetInput
		IsOutput func(tb testing.TB, out *svc.DownloadDatasetOutput) bool
		IsErr    func(error) bool
	}{
		{
			Name:    "when dataset doesnt exist it should return an error",
			Timeout: time.Second * 5,
			Input:   &svc.DownloadDatasetInput{Name: "my-dataset"},
			IsErr:   kubevisor.IsNotExistsErr,
			IsOutput: func(t testing.TB, out *svc.DownloadDatasetOutput) bool {
				return true
			},
		},
		{
			Name:     "when one dataset has been uploaded it should be available for download",
			Timeout:  time.Minute,
			Datasets: []*svc.UploadDatasetInput{{Name: "my-dataset", Dir: "/tmp"}},
			Input:    &svc.DownloadDatasetInput{Name: "my-datasets"},
			IsErr:    nil,
			IsOutput: func(t testing.TB, out *svc.DownloadDatasetOutput) bool {
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
				_, err := kube.UploadDataset(ctx, datasets)
				ok(t, err)
			}

			out, err := kube.DownloadDataset(ctx, c.Input)
			if c.IsErr != nil { //if c.IsErr is nil we dont care about errors
				assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			}

			if c.IsOutput != nil {
				c.IsOutput(t, out)
			}
		})
	}
}
