package svc_test

import (
	"context"
	"testing"
	"time"

	"github.com/nerdalize/nerd/svc"
)

func TestUpdateSecret(t *testing.T) {
	di, clean := testDI(t)
	defer clean()

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	kube := svc.NewKube(di)
	out, err := kube.CreateSecret(ctx, &svc.CreateSecretInput{
		Image:    "smoketest",
		Project:  "nerdalize",
		Registry: "quay.io",
		Username: "test",
		Password: "test",
	})
	ok(t, err)

	_, err = kube.UpdateSecret(ctx, &svc.UpdateSecretInput{
		Name:     out.Name,
		Username: "newtest",
		Password: "newtest",
	})
	ok(t, err)
}
