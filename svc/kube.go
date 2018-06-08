package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"
)

//Kube interacts with the kubernetes backend
type Kube struct {
	visor *kubevisor.Visor
	val   Validator
	logs  Logger
}

//NewKube will setup the Kubernetes service
func NewKube(di DI) (k *Kube) {
	k = &Kube{
		visor: kubevisor.NewVisor(di.Namespace(), "", di.Kube(), di.Crd(), di.APIExt(), di.Logger()),
		val:   di.Validator(),
		logs:  di.Logger(),
	}

	return k
}

func (k *Kube) checkInput(ctx context.Context, in interface{}) (err error) {
	err = k.val.StructCtx(ctx, in)
	if err != nil {
		return errValidation{err}
	}

	return nil
}
