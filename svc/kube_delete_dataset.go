package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"
)

//DeleteDatasetInput is the input to DeleteDataset
type DeleteDatasetInput struct {
	Name string `validate:"min=1,printascii"`
}

//DeleteDatasetOutput is the output to DeleteDataset
type DeleteDatasetOutput struct{}

//DeleteDataset will create a dataset on kubernetes
func (k *Kube) DeleteDataset(ctx context.Context, in *DeleteDatasetInput) (out *DeleteDatasetOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	err = k.visor.DeleteResource(ctx, kubevisor.ResourceTypeDatasets, in.Name)
	if err != nil {
		return nil, err
	}

	return &DeleteDatasetOutput{}, nil
}
