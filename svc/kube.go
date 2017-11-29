package svc

import (
	"k8s.io/client-go/kubernetes"
)

//Kube interacts with the kubernetes backend
type Kube struct {
	ns  string
	api kubernetes.Interface
	val Validator
}

//NewKube will setup the Kubernetes service
func NewKube(di DI, ns string) (k *Kube, err error) {
	k = &Kube{
		ns:  ns,
		api: di.Kube(),
		val: di.Validator(),
	}

	return k, nil
}
