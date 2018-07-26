package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"
)

//DeleteFileSystemInput is the input to DeleteFileSystem
type DeleteFileSystemInput struct {
	Name string `validate:"min=1,printascii"`
}

//DeleteFileSystemOutput is the output to DeleteFileSystem
type DeleteFileSystemOutput struct{}

//DeleteFileSystem will create a dataset on kubernetes
func (k *Kube) DeleteFileSystem(ctx context.Context, in *DeleteFileSystemInput) (out *DeleteFileSystemOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	err = k.visor.DeleteResource(ctx, kubevisor.ResourceTypePersistentVolumeClaims, in.Name)
	if err != nil {
		return nil, err
	}

	return &DeleteFileSystemOutput{}, nil
}
