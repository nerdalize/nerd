package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"
)

//DeleteSecretInput is the input to DeleteSecret
type DeleteSecretInput struct {
	Name string `validate:"min=1,printascii"`
}

//DeleteSecretOutput is the output to DeleteSecret
type DeleteSecretOutput struct{}

//DeleteSecret will create a dataset on kubernetes
func (k *Kube) DeleteSecret(ctx context.Context, in *DeleteSecretInput) (out *DeleteSecretOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	err = k.visor.DeleteResource(ctx, kubevisor.ResourceTypeSecrets, in.Name)
	if err != nil {
		return nil, err
	}

	return &DeleteSecretOutput{}, nil
}
