package svc_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nerdalize/nerd/svc"
)

func TestUpdateDataset(t *testing.T) {
	di, clean := testDI(t)
	defer clean()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	kube := svc.NewKube(di)
	out, err := kube.CreateDataset(ctx, &svc.CreateDatasetInput{Name: "my-dataset", Bucket: "bogus", Key: "my-key"})
	ok(t, err)

	_, err = kube.UpdateDataset(ctx, &svc.UpdateDatasetInput{
		Name:       out.Name,
		InputFor:   "j-123abc",
		OutputFrom: "j-456def",
	})
	ok(t, err)

	o, err := kube.GetDataset(ctx, &svc.GetDatasetInput{Name: out.Name})
	ok(t, err)
	assert(t, strings.Contains(strings.Join(o.InputFor, ""), "j-123abc"), "expected dataset to be up to date")
	assert(t, strings.Contains(strings.Join(o.OutputFrom, ""), "j-456def"), "expected dataset to be up to date and to contain job info for output section")
}
