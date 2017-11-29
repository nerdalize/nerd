package svc

import (
	"k8s.io/client-go/kubernetes"
)

//Kube interacts with the kubernetes backend
type Kube struct {
	api kubernetes.Interface
}

//NewKube will setup the Kubernetes service
func NewKube(di DI) (k *Kube, err error) {
	k = &Kube{
		api: di.Kube(),
	}

	// kcfg, err := clientcmd.BuildConfigFromFlags("", confPath)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed to build kubernetes configuration")
	// }
	//
	// k = &Kube{}
	// k.api, err = kubernetes.NewForConfig(kcfg)
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failed initiate kubernetes client set")
	// }

	return k, nil
}
