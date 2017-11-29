package svc_test

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/nerdalize/nerd/svc"
)

func TestRunJob(t *testing.T) {
	for _, c := range []struct {
		Name   string
		Ctx    context.Context
		Input  *svc.RunJobInput
		Output *svc.RunJobOutput
		IsErr  func(error) bool
	}{
		{
			Name:   "when a zero value input is provided it should return a no input error",
			Ctx:    context.Background(),
			Input:  nil,
			Output: nil,
			IsErr:  svc.IsNoInputErr,
		},
		// {
		// 	Name:   "when no namespace is available, it should return a specific error",
		// 	Ctx:    context.Background(),
		// 	Input:  nil,
		// 	Output: nil,
		// 	Error:  nil,
		// },
	} {
		t.Run(c.Name, func(t *testing.T) {
			di := testDI(t)
			kube, err := svc.NewKube(di)
			ok(t, err)

			out, err := kube.RunJob(c.Ctx, c.Input)
			assert(t, c.IsErr(err), fmt.Sprintf("unexpected '%#v' to match: %#v", err, runtime.FuncForPC(reflect.ValueOf(c.IsErr).Pointer()).Name()))
			equals(t, c.Output, out)
		})
	}

	// t.Run("when no namespace is available, it should return a specific error", func(t *testing.T) {
	//
	// })
	//
	// t.Run("name", func(t *testing.T) {
	//
	// })

	//@TODO test the case in which no namespace is available
}
