package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"
)

//DeleteJobInput is the input to DeleteJob
type DeleteJobInput struct {
	Name string `validate:"min=1,printascii"`
}

//DeleteJobOutput is the output to DeleteJob
type DeleteJobOutput struct{}

//DeleteJob will create a job on kubernetes
func (k *Kube) DeleteJob(ctx context.Context, in *DeleteJobInput) (out *DeleteJobOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	err = k.visor.DeleteResource(ctx, kubevisor.ResourceTypeJobs, in.Name)
	if err != nil {
		return nil, err
	}

	return &DeleteJobOutput{}, nil
}
