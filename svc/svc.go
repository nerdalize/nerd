package svc

import "k8s.io/client-go/kubernetes"

//DI provides dependencies for our services
type DI interface {
	Kube() kubernetes.Interface
}
