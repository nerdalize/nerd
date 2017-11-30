package svc

import (
	"context"

	"k8s.io/client-go/kubernetes"
)

//Validator describes the validation dependency we require
type Validator interface {
	StructCtx(ctx context.Context, s interface{}) (err error)
}

//Logger describes the logging dependency the services require
type Logger interface {
	Debugf(format string, args ...interface{})
}

//DI provides dependencies for our services
type DI interface {
	Kube() kubernetes.Interface
	Validator() Validator
	Logger() Logger
}
