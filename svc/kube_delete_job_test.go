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

func TestDeleteJob(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Jobs     []*svc.RunJobInput
		Input    *svc.DeleteJobInput
		Output   *svc.DeleteJobOutput
		Listing  *svc.ListJobsOutput
		IsOutput func(tb testing.TB, out *svc.DeleteJobOutput, list *svc.ListJobsOutput)
		IsErr    func(error) bool
	}{
		{
			Name:    "when no name is provided it should provide a validation error",
			Timeout: time.Second * 5,
			Input:   &svc.DeleteJobInput{},
			Output:  &svc.DeleteJobOutput{},
			IsErr:   svc.IsValidationErr,
		},
		{
			Name:    "when an existing job is delete it should be marked for garbage collection",
			Timeout: time.Second * 5,
			Jobs:    []*svc.RunJobInput{{Image: "nginx", Name: "my-job"}},
			Input:   &svc.DeleteJobInput{Name: "my-job"},
			Output:  &svc.DeleteJobOutput{},
			IsErr:   isNilErr,
			IsOutput: func(t testing.TB, out *svc.DeleteJobOutput, list *svc.ListJobsOutput) {
				assert(t, len(list.Items) == 1, "job should still be there")
				assert(t, !list.Items[0].DeletedAt.IsZero(), "delete at should not be zero")
			},
		},
		{
			Name:    "when a non-existing job is delete it should return NotExists error",
			Timeout: time.Second * 5,
			Jobs:    []*svc.RunJobInput{{Image: "nginx", Name: "my-job"}},
			Input:   &svc.DeleteJobInput{Name: "foo"},
			Output:  &svc.DeleteJobOutput{},
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
			for _, job := range c.Jobs {
				_, err := kube.RunJob(ctx, job)
				ok(t, err)
			}

			out, err := kube.DeleteJob(ctx, c.Input)
			if c.IsErr != nil {
				assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			}

			list, err := kube.ListJobs(ctx, &svc.ListJobsInput{})
			ok(t, err)

			if c.IsOutput != nil {
				c.IsOutput(t, out, list)
			}
		})
	}
}
