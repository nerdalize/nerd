package svc_test

import (
	"context"
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

	_, err = kube.UpdateDataset(ctx, &svc.UpdateDatasetInput{Name: out.Name, Input: "j-123abc"})
	ok(t, err)

	o, err := kube.GetDataset(ctx, &svc.GetDatasetInput{Name: out.Name})
	ok(t, err)
	assert(t, o.Input == "j-123abc", "expected dataset to be up to date")
}
