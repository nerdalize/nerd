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

	newSize := uint64(1337)
	_, err = kube.UpdateDataset(ctx, &svc.UpdateDatasetInput{
		Name:       out.Name,
		InputFor:   "j-123abc",
		OutputFrom: "j-456def",
		Size:       &newSize,
	})
	ok(t, err)

	o, err := kube.GetDataset(ctx, &svc.GetDatasetInput{Name: out.Name})
	ok(t, err)
	assert(t, strings.Contains(strings.Join(o.InputFor, ""), "j-123abc"), "expected dataset to be up to date")
	assert(t, strings.Contains(strings.Join(o.OutputFrom, ""), "j-456def"), "expected dataset to be up to date and to contain job info for output section")
	assert(t, o.Size == 1337, "expected dataset to be up to date and contain new size")

	//Check if the output remains the same when not specifying any changes
	_, err = kube.UpdateDataset(ctx, &svc.UpdateDatasetInput{
		Name:       out.Name,
	})
	ok(t, err)

	o2, err := kube.GetDataset(ctx, &svc.GetDatasetInput{Name: out.Name})
	ok(t, err)
	equals(t, o.Size, o2.Size)
	equals(t, o.InputFor, o2.InputFor)
	equals(t, o.OutputFrom, o2.OutputFrom)
}
