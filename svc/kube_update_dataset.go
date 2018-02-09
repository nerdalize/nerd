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
	Size       *uint64
	InputFor   string
	OutputFrom string
}

// UpdateDatasetOutput is the output for UpdateDataset
type UpdateDatasetOutput struct {
	Name string
}

// UpdateDataset will update a dataset resource.
// Fields that can be updated: name, input, output and size. Input and output are the jobs the dataset is used for or coming from.
// Size is only updated if it's 0 or higher
func (k *Kube) UpdateDataset(ctx context.Context, in *UpdateDatasetInput) (out *UpdateDatasetOutput, err error) {
	dataset := &datasetsv1.Dataset{}
	err = k.visor.GetResource(ctx, kubevisor.ResourceTypeDatasets, dataset, in.Name)
	if err != nil {
		return nil, err
	}

	if in.NewName != "" {
		dataset.SetName(in.NewName)
	}
	if in.Size != nil {
		dataset.Spec.Size = *in.Size
	}
	if in.InputFor != "" {
		dataset.Spec.InputFor = append(dataset.Spec.InputFor, in.InputFor)
	}
	if in.OutputFrom != "" {
		dataset.Spec.OutputFrom = append(dataset.Spec.OutputFrom, in.OutputFrom)
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
