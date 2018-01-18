package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"

	datasetsv1 "github.com/nerdalize/nerd/crd/pkg/apis/stable.nerdalize.com/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//UploadDatasetInput is the input to UploadDataset
type UploadDatasetInput struct {
	Dir  string `validate:"min=1"`
	Name string `validate:"printascii"`
}

//UploadDatasetOutput is the output to UploadDataset
type UploadDatasetOutput struct {
	Name string
}

//UploadDataset will create a dataset on kubernetes
func (k *Kube) UploadDataset(ctx context.Context, in *UploadDatasetInput) (out *UploadDatasetOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	//@TODO Upload dir to S3 bucket

	dataset := &datasetsv1.Dataset{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: datasetsv1.DatasetSpec{
			Bucket: "to-be-determined",
			Key:    "tbd",
		},
	}

	err = k.visor.CreateResource(ctx, kubevisor.ResourceTypeDatasets, dataset, in.Name)
	if err != nil {
		return nil, err
	}

	return &UploadDatasetOutput{
		Name: dataset.Name,
	}, nil
}
