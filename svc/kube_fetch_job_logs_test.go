package svc_test

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/nerdalize/nerd/pkg/kubevisor"
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
			IsErr:   kubevisor.IsNotExistsErr,
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
		{
			Name:    "tail option should allow for limiting the nr of lines to return",
			Timeout: time.Minute,
			Jobs:    []*svc.RunJobInput{{Image: "hello-world", Name: "my-job"}},
			Input:   &svc.FetchJobLogsInput{Name: "my-job", Tail: 3},
			IsErr:   nil,
			IsOutput: func(t testing.TB, out *svc.FetchJobLogsOutput) bool {
				if out == nil || len(out.Data) < 1 {
					return false
				}

				assert(t, !bytes.Contains(out.Data, []byte("Hello from Docker")), "logs should not contain the data before the tail")
				assert(t, bytes.Contains(out.Data, []byte("more examples and ideas")), "logs should contain the data after the tail")
				return true
			},
		},

		//@TODO find a way to not be dependant on a specific key to be present on s3
		// {
		// 	Name:    "when one job with a volume is run it should return logs indicating the download",
		// 	Timeout: time.Minute,
		// 	Jobs: []*svc.RunJobInput{{
		// 		Image: "alpine",
		// 		Name:  "my-job",
		// 		Args:  []string{"ls", "-la", "/input"},
		// 		Volumes: []svc.JobVolume{
		// 			{
		// 				Path:   "/input",
		// 				Type:   svc.JobVolumeTypeInput,
		// 				Bucket: "nlz-datasets-dev",
		// 				Key:    "959d240b-4abb-43ba-b253-98b024541e09.zip",
		// 			},
		// 		},
		// 	}},
		// 	Input: &svc.FetchJobLogsInput{Name: "my-job"},
		// 	IsErr: nil,
		// 	IsOutput: func(t testing.TB, out *svc.FetchJobLogsOutput) bool {
		// 		if out == nil || len(out.Data) < 1 {
		// 			return false
		// 		}
		//
		// 		assert(t, bytes.Contains(out.Data, []byte("westerscheldt_0001_19920831_020000")), "logs should contain the data we expect")
		// 		return true
		// 	},
		// },

		//@TODO add a test with multiple jobs, make sure logs are returned from earlier jobs
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
