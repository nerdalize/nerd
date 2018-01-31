package svc

import (
	"context"
	"log"

	"github.com/nerdalize/nerd/pkg/kubevisor"

	datasetsv1 "github.com/nerdalize/nerd/crd/pkg/apis/stable.nerdalize.com/v1"
)

// UpdateDatasetInput is the input for UpdateDataset
type UpdateDatasetInput struct {
	Name       string `validate:"printascii"`
	NewName    string `validate:"printascii"`
	InputFor   string
	OutputFrom string
}

// UpdateDatasetOutput is the output for UpdateDataset
type UpdateDatasetOutput struct {
	Name string
}

// UpdateDataset will update a dataset resource.
// Fields that can be updated: name, input, and output. Input and output are the jobs the dataset is used for or coming from.
func (k *Kube) UpdateDataset(ctx context.Context, in *UpdateDatasetInput) (out *UpdateDatasetOutput, err error) {
	dataset := &datasetsv1.Dataset{}
	err = k.visor.GetResource(ctx, kubevisor.ResourceTypeDatasets, dataset, in.Name)
	if err != nil {
		return nil, err
	}

	if in.NewName != "" {
		dataset.SetName(in.NewName)
	}
	dataset.Spec = datasetsv1.DatasetSpec{
		InputFor:   in.InputFor,
		OutputFrom: in.OutputFrom,
	}

	err = k.visor.UpdateResource(ctx, kubevisor.ResourceTypeDatasets, dataset, in.Name)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &UpdateDatasetOutput{
		Name: dataset.Name,
	}, nil
}
