package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"

	datasetsv1 "github.com/nerdalize/nerd/crd/pkg/apis/stable.nerdalize.com/v1"
	"github.com/nerdalize/nerd/pkg/transfer/archiver"
	"github.com/nerdalize/nerd/pkg/transfer/store"
)

//GetDatasetInput is the input to GetDataset
type GetDatasetInput struct {
	Name string `validate:"printascii"`
}

//GetDatasetOutput is the output to GetDataset
type GetDatasetOutput struct {
	Name string
	Size uint64

	InputFor   []string
	OutputFrom []string

	StoreOptions    transferstore.StoreOptions
	ArchiverOptions transferarchiver.ArchiverOptions
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

	return GetDatasetOutputFromSpec(dataset), nil
}

//GetDatasetOutputFromSpec allows easy output creation from dataset
func GetDatasetOutputFromSpec(dataset *datasetsv1.Dataset) *GetDatasetOutput {
	return &GetDatasetOutput{
		Name:            dataset.Name,
		Size:            dataset.Spec.Size,
		InputFor:        dataset.Spec.InputFor,
		OutputFrom:      dataset.Spec.OutputFrom,
		StoreOptions:    dataset.Spec.StoreOptions,
		ArchiverOptions: dataset.Spec.ArchiverOptions,
	}
}
