package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"

	"k8s.io/api/core/v1"
)

//GetFileSystemInput is the input to GetFileSystem
type GetFileSystemInput struct {
	Name string `validate:"printascii"`
}

//SambaMountInfo provides the info to set up samba access to the file system contents
type SambaMountInfo struct {
	Path string
}

//GetFileSystemOutput is the output to GetFileSystem
type GetFileSystemOutput struct {
	Name string
	VolumeName string
}

//GetFileSystem will retrieve a filesystem from kubernetes
func (k *Kube) GetFileSystem(ctx context.Context, in *GetFileSystemInput) (out *GetFileSystemOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	pvc := &v1.PersistentVolumeClaim{}
	err = k.visor.GetResource(ctx, kubevisor.ResourceTypePersistentVolumeClaims, pvc, in.Name)
	if err != nil {
		return nil, err
	}

	return GetFileSystemOutputFromSpec(pvc), nil
}

//GetFileSystemOutputFromSpec allows easy output creation from filesystem
func GetFileSystemOutputFromSpec(pvc *v1.PersistentVolumeClaim) *GetFileSystemOutput {
	return &GetFileSystemOutput{
		Name:       pvc.Name,
		VolumeName: pvc.Spec.VolumeName,
	}
}
