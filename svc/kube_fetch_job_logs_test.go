package svc_test

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/nerdalize/nerd/svc"
)

func TestFetchJobLogs(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Jobs     []*svc.RunJobInput
		Input    *svc.FetchJobLogsInput
		IsOutput func(tb testing.TB, out *svc.FetchJobLogsOutput) bool
		IsErr    func(error) bool
	}{
		{
			Name:    "when job doesnt exist it should that there were no logs available",
			Timeout: time.Second * 5,
			Input:   &svc.FetchJobLogsInput{Name: "my-job"},
			IsErr:   svc.IsNoLogsErr,
			IsOutput: func(t testing.TB, out *svc.FetchJobLogsOutput) bool {
				return true
			},
		},
		{
			Name:    "when one correct job was run it should eventually return logs",
			Timeout: time.Minute,
			Jobs:    []*svc.RunJobInput{{Image: "hello-world", Name: "my-job"}},
			Input:   &svc.FetchJobLogsInput{Name: "my-job"},
			IsErr:   nil,
			IsOutput: func(t testing.TB, out *svc.FetchJobLogsOutput) bool {
				if out == nil || len(out.Data) < 1 {
					return false
				}

				assert(t, bytes.Contains(out.Data, []byte("Hello from Docker")), "logs should contain the data we expect")
				return true
			},
		},
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

			out, err := kube.FetchJobLogs(ctx, c.Input)
			if c.IsErr != nil { //if c.IsErr is nil we dont care about errors
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
				t.Logf("retrying logs in %s...", d)
				<-time.After(d)

				out, err = kube.FetchJobLogs(ctx, c.Input)
				if err != nil && c.IsErr != nil { //if c.IsErr is nil we dont care about errors
					t.Fatalf("failed to list jobs during retry: %v", err)
				}
			}
		})
	}
}
