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

func TestDeleteDataset(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Datasets []*svc.CreateDatasetInput
		Input    *svc.DeleteDatasetInput
		Output   *svc.DeleteDatasetOutput
		Listing  *svc.ListDatasetsOutput
		IsOutput func(tb testing.TB, out *svc.DeleteDatasetOutput, l *svc.ListDatasetsOutput)
		IsErr    func(error) bool
	}{
		{
			Name:    "when no name is provided it should provide a validation error",
			Timeout: time.Second * 5,
			Input:   &svc.DeleteDatasetInput{},
			Output:  &svc.DeleteDatasetOutput{},
			IsErr:   svc.IsValidationErr,
		},
		{
			Name:    "when a non-existing dataset is deleted it should return NotExists error",
			Timeout: time.Second * 5,
			Input:   &svc.DeleteDatasetInput{Name: "foo"},
			Output:  &svc.DeleteDatasetOutput{},
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
			for _, dataset := range c.Datasets {
				_, err := kube.CreateDataset(ctx, dataset)
				ok(t, err)
			}

			out, err := kube.DeleteDataset(ctx, c.Input)
			if c.IsErr != nil {
				assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			}

			list, err := kube.ListDatasets(ctx, &svc.ListDatasetsInput{})
			ok(t, err)

			if c.IsOutput != nil {
				c.IsOutput(t, out, list)
			}
		})
	}
}
