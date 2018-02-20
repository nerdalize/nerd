package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"

	datasetsv1 "github.com/nerdalize/nerd/crd/pkg/apis/stable.nerdalize.com/v1"
	"github.com/nerdalize/nerd/pkg/transfer/archiver"
	"github.com/nerdalize/nerd/pkg/transfer/store"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//CreateDatasetInput is the input to CreateDataset
type CreateDatasetInput struct {
	Bucket string //@TODO deprecate for more flexible options
	Key    string //@TODO deprecate for more flexible options

	Name string `validate:"printascii"`
	Size uint64

	//@TODO add concrete type
	StoreOptions    transferstore.StoreOptions       `validate:"required"`
	ArchiverOptions transferarchiver.ArchiverOptions `validate:"required"`

	// StoreType    string `validate:"min=1"`
	// ArchiverType string `validate:"min=1"`
	// Options      map[string]string
}

//CreateDatasetOutput is the output to CreateDataset
type CreateDatasetOutput struct {
	Name string
}

//CreateDataset will create a dataset on kubernetes
func (k *Kube) CreateDataset(ctx context.Context, in *CreateDatasetInput) (out *CreateDatasetOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	dataset := &datasetsv1.Dataset{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: datasetsv1.DatasetSpec{
			Size:   in.Size,
			Bucket: in.Bucket, //@TODO deprecate
			Key:    in.Key,    //@TODO deprecate

			StoreOptions:    in.StoreOptions,
			ArchiverOptions: in.ArchiverOptions,

			// StoreType:    in.StoreType,
			// ArchiverType: in.ArchiverType,
			// Options:      in.Options,
		},
	}

	err = k.visor.CreateResource(ctx, kubevisor.ResourceTypeDatasets, dataset, in.Name)
	if err != nil {
		return nil, err
	}

	return &CreateDatasetOutput{
		Name: dataset.Name,
	}, nil
}
