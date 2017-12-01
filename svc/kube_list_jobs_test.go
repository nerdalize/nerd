package svc_test

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/nerdalize/nerd/svc"
)

func TestListJobs(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Jobs     []*svc.RunJobInput
		Input    *svc.ListJobsInput
		Output   *svc.ListJobsOutput
		IsOutput func(tb testing.TB, out *svc.ListJobsOutput)
		IsErr    func(error) bool
	}{
		{
			Name:    "when no jobs have run the output should be empty",
			Timeout: time.Second * 5,
			Input:   &svc.ListJobsInput{},
			Output:  &svc.ListJobsOutput{},
			IsErr:   isNilErr,
		},
		{
			Name:    "when one jobs has run the output should not be empty",
			Timeout: time.Second * 5,
			Jobs:    []*svc.RunJobInput{{Image: "nginx", Name: "my-job"}},
			Input:   &svc.ListJobsInput{},
			Output:  &svc.ListJobsOutput{},
			IsErr:   isNilErr,
			IsOutput: func(t testing.TB, out *svc.ListJobsOutput) {
				assert(t, len(out.Items) == 1, "expected one job to be listed")
				assert(t, out.Items[0].Name == "my-job", "expected job name to be equal to what was provided")
				assert(t, out.Items[0].Image == "nginx", "expected image name to be equal to what was provided")
			},
		},
		//@TODO test if non-cli jobs show up in kubernetes
	} {
		t.Run(c.Name, func(t *testing.T) {
			di := testDI(t)
			ns, clean := testNamespace(t, di.Kube())
			defer clean()

			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, c.Timeout)
			defer cancel()

			kube := svc.NewKube(di, ns)
			for _, job := range c.Jobs {
				_, err := kube.RunJob(ctx, job)
				ok(t, err)
			}

			out, err := kube.ListJobs(ctx, c.Input)
			if c.IsErr != nil {
				assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			}

			if c.IsOutput != nil {
				c.IsOutput(t, out)
			}
		})
	}
}
