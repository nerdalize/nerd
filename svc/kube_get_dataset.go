package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"

	datasetsv1 "github.com/nerdalize/nerd/crd/pkg/apis/stable.nerdalize.com/v1"
)

//GetDatasetInput is the input to GetDataset
type GetDatasetInput struct {
	Name string `validate:"printascii"`
}

//GetDatasetOutput is the output to GetDataset
type GetDatasetOutput struct {
	Name       string
	Bucket     string
	Key        string
	Size       uint64
	InputFor   string
	OutputFrom string
}

//GetDataset will create a dataset on kubernetes
func (k *Kube) GetDataset(ctx context.Context, in *GetDatasetInput) (out *GetDatasetOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	dataset := &datasetsv1.Dataset{}
	err = k.visor.GetResource(ctx, kubevisor.ResourceTypeDatasets, dataset, in.Name)
	if err != nil {
		return nil, err
	}

	return &GetDatasetOutput{
		Name:       dataset.Name,
		Size:       dataset.Spec.Size,
		Bucket:     dataset.Spec.Bucket,
		Key:        dataset.Spec.Key,
		InputFor:   dataset.Spec.InputFor,
		OutputFrom: dataset.Spec.OutputFrom,
	}, nil
}
