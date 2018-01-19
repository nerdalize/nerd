package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"

	datasetsv1 "github.com/nerdalize/nerd/crd/pkg/apis/stable.nerdalize.com/v1"
)

//DownloadDatasetInput is the input to DownloadDataset
type DownloadDatasetInput struct {
	JobInput  string
	JobOutput string
	Dest      string `validate:"min=1"`
	Name      string `validate:"printascii"`
}

//DownloadDatasetOutput is the output to DownloadDataset
type DownloadDatasetOutput struct {
	Name string
}

//DownloadDataset will create a dataset on kubernetes
func (k *Kube) DownloadDataset(ctx context.Context, in *DownloadDatasetInput) (out *DownloadDatasetOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	dataset := &datasetsv1.Dataset{}

	err = k.visor.GetResource(ctx, kubevisor.ResourceTypeDatasets, dataset, in.Name)
	if err != nil {
		return nil, err
	}

	return &DownloadDatasetOutput{
		Name: dataset.Name,
	}, nil
}
