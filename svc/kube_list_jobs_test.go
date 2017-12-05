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

var (
	noBackoffLimit = int32(0)
)

func TestListJobs(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Jobs     []*svc.RunJobInput
		Input    *svc.ListJobsInput
		IsOutput func(tb testing.TB, out *svc.ListJobsOutput) bool
		IsErr    func(error) bool
	}{
		{
			Name:    "when no jobs have run the output should be empty",
			Timeout: time.Second * 5,
			Input:   &svc.ListJobsInput{},
			IsErr:   isNilErr,
			IsOutput: func(t testing.TB, out *svc.ListJobsOutput) bool {
				assert(t, len(out.Items) == 0, "expected zero jobs to be listed")
				return true
			},
		},
		{
			Name:    "when one correct job was run it should eventually be started",
			Timeout: time.Second * 5,
			Jobs:    []*svc.RunJobInput{{Image: "nginx", Name: "my-job"}},
			Input:   &svc.ListJobsInput{},
			IsErr:   isNilErr,
			IsOutput: func(t testing.TB, out *svc.ListJobsOutput) bool {
				assert(t, len(out.Items) == 1, "expected one job to be listed")
				assert(t, !out.Items[0].CreatedAt.IsZero(), "created at time should not be zero")

				assert(t, out.Items[0].Name == "my-job", "expected job name to be equal to what was provided")
				assert(t, out.Items[0].Image == "nginx", "expected image name to be equal to what was provided")
				if out.Items[0].ActiveAt.IsZero() {
					return false //should eventually start
				}

				assert(t, out.Items[0].ActiveAt.Sub(time.Now()) < time.Minute, "started time should be recent")
				return true
			},
		},
		{
			Name:    "when a short job is started it should show completed at some point",
			Timeout: time.Second * 20,
			Jobs:    []*svc.RunJobInput{{Image: "hello-world", Name: "my-job"}},
			Input:   &svc.ListJobsInput{},
			IsErr:   isNilErr,
			IsOutput: func(t testing.TB, out *svc.ListJobsOutput) bool {
				assert(t, len(out.Items) == 1, "expected one job to be listed")
				if out.Items[0].CompletedAt.IsZero() {
					return false
				}

				assert(t, out.Items[0].CompletedAt.Sub(time.Now()) < time.Minute, "completed time should be recent")
				return true
			},
		},
		{
			Name:    "when job is run with the wrong image it should still become active",
			Timeout: time.Minute,
			Jobs:    []*svc.RunJobInput{{Image: "there-is-no-image-called-this", Name: "invalid-image-job"}},
			Input:   &svc.ListJobsInput{},
			IsErr:   isNilErr,
			IsOutput: func(t testing.TB, out *svc.ListJobsOutput) bool {
				assert(t, len(out.Items) == 1, "expected one job to be listed")
				if out.Items[0].ActiveAt.IsZero() {
					return false //should eventually be active
				}

				assert(t, out.Items[0].ActiveAt.Sub(time.Now()) < time.Minute, "started time should be recent")
				return true
			},
		},
		{
			Name:    "when a job is run that fails it should eventually become failed",
			Timeout: time.Minute,
			Jobs:    []*svc.RunJobInput{{Image: "vmarmol/false", Name: "failing-job", BackoffLimit: &noBackoffLimit}},
			Input:   &svc.ListJobsInput{},
			IsErr:   isNilErr,
			IsOutput: func(t testing.TB, out *svc.ListJobsOutput) bool {
				assert(t, len(out.Items) == 1, "expected one job to be listed")
				if out.Items[0].FailedAt.IsZero() {
					return false //should eventually fail
				}

				assert(t, out.Items[0].FailedAt.Sub(time.Now()) < time.Minute, "started time should be recent")
				return true
			},
		},

		//@TODO when a job is scaled to 0, find out the status shows it (if parralism is 0, it is stopped)
	} {
		t.Run(c.Name, func(t *testing.T) {
			if c.Timeout > time.Second*5 && testing.Short() {
				t.Skipf("skipping long test with contex timeout: %s", c.Timeout)
			}

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

			if c.IsOutput == nil {
				return //no output testing
			}

			for {
				if c.IsOutput(t, out) {
					break
				}

				d := time.Second
				t.Logf("retrying listing in %s...", d)
				<-time.After(d)

				out, err = kube.ListJobs(ctx, c.Input)
				if err != nil {
					t.Fatalf("failed to list jobs during retry: %v", err)
				}
			}
		})
	}
}
