package svc

import (
	"context"

	"k8s.io/client-go/kubernetes"
)

//Validator describes the validation dependency we require
type Validator interface {
	StructCtx(ctx context.Context, s interface{}) (err error)
}

//DI provides dependencies for our services
type DI interface {
	Kube() kubernetes.Interface
	Validator() Validator
}
