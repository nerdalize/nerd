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
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRunJob(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Input    *svc.RunJobInput
		Output   *svc.RunJobOutput
		IsOutput func(tb testing.TB, out *svc.RunJobOutput)
		IsErr    func(error) bool
	}{
		{
			Name:    "when a zero value input is provided it should return a validation error",
			Timeout: time.Second * 5,
			Input:   nil,
			IsErr:   svc.IsValidationErr,
			IsOutput: func(t testing.TB, out *svc.RunJobOutput) {
				assert(t, out == nil, "output should be nil")
			},
		},
		{
			Name:    "when input is provided that is invalid it should return a validation error",
			Timeout: time.Second * 5,
			Input:   &svc.RunJobInput{},
			IsErr:   svc.IsValidationErr,
			IsOutput: func(t testing.TB, out *svc.RunJobOutput) {
				assert(t, out == nil, "output should be nil")
			},
		},
		{
			Name:    "when a job is started with just an image it should generate a name and return it",
			Timeout: time.Second * 5,
			Input:   &svc.RunJobInput{Image: "hello-world"},
			IsErr:   isNilErr,
			IsOutput: func(t testing.TB, out *svc.RunJobOutput) {
				assert(t, out != nil, "output should not be nil")
				assert(t, regexp.MustCompile(`^j-.+$`).MatchString(out.Name), "name should have a prefix but not be empty after the prefix")
			},
		},
		{
			Name:    "when a job is started with a very short deadline it should return a specific error",
			Timeout: time.Millisecond,
			Input:   &svc.RunJobInput{Image: "hello-world"},
			IsErr:   kubevisor.IsDeadlineErr,
		},
		{
			Name:    "when a job is started with an invalid name it should return a invalid name error",
			Timeout: time.Second * 5,
			Input:   &svc.RunJobInput{Image: "hello-world", Name: "my-name-"},
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
			out, err := kube.RunJob(ctx, c.Input)
			if c.IsErr != nil {
				assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			}

			if c.IsOutput != nil {
				c.IsOutput(t, out)
			}
		})
	}
}

func TestRunJobTemplate(t *testing.T) {
	for _, c := range []struct {
		Name     string
		Timeout  time.Duration
		Input    *svc.RunJobInput
		IsOutput func(tb testing.TB, out *batchv1.Job) bool
	}{
		{
			Name:    "when args are provided they should be found in the pod template",
			Timeout: time.Second * 10,
			Input:   &svc.RunJobInput{Image: "nginx", Name: "my-args", Args: []string{"--port=8080"}},
			IsOutput: func(t testing.TB, out *batchv1.Job) bool {
				if len(out.Spec.Template.Spec.Containers) < 1 {
					return false
				}
				assert(t, len(out.Spec.Template.Spec.Containers[0].Args) == 1, "there should be at least one arg")
				assert(t, out.Spec.Template.Spec.Containers[0].Args[0] == "--port=8080", "arg should be the one given in input")
				return true
			},
		},
		{
			Name:    "when env vars are provided they should be found in the job template",
			Timeout: time.Second * 10,
			Input:   &svc.RunJobInput{Image: "nginx", Name: "my-env", Env: map[string]string{"TEST": "xyz"}},
			IsOutput: func(t testing.TB, out *batchv1.Job) bool {
				if len(out.Spec.Template.Spec.Containers) < 1 {
					return false
				}
				assert(t, len(out.Spec.Template.Spec.Containers[0].Env) == 1, "there should be at least one environment variable")
				assert(t, out.Spec.Template.Spec.Containers[0].Env[0].Name == "TEST", "env var name should be the one given in input")
				assert(t, out.Spec.Template.Spec.Containers[0].Env[0].Value == "xyz", "env var value should be the one given in input")
				return true
			},
		},
	} {
		t.Run(c.Name, func(t *testing.T) {
			di, clean := testDI(t)
			defer clean()

			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, c.Timeout)
			defer cancel()

			kube := svc.NewKube(di)
			o, err := kube.RunJob(ctx, c.Input)
			ok(t, err)
			assert(t, o != nil, "expected RunJob return to be not nil")

			k := di.Kube()
			out, err := k.BatchV1().Jobs(di.Namespace()).Get("nlz-nerd"+o.Name, metav1.GetOptions{})
			if c.IsOutput == nil {
				return
			}
			assert(t, c.IsOutput(t, out), "should return true")
		})
	}
}

func TestRunJobWithoutNamespace(t *testing.T) {
	di := testDIWithoutNamespace(t)

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	kube := svc.NewKube(di)
	_, err := kube.RunJob(ctx, &svc.RunJobInput{Image: "hello-world", Name: "my-job"})
	assert(t, kubevisor.IsNamespaceNotExistsErr(err), "expected error to be namespace doesn't exist")
}

func TestRunJobWithNameThatAlreadyExists(t *testing.T) {
	di, clean := testDI(t)
	defer clean()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	kube := svc.NewKube(di)
	out, err := kube.RunJob(ctx, &svc.RunJobInput{Image: "hello-world", Name: "my-job"})
	ok(t, err)

	_, err = kube.RunJob(ctx, &svc.RunJobInput{Image: "hello-world", Name: out.Name})
	assert(t, kubevisor.IsAlreadyExistsErr(err), "expected error to be already exists")
}
