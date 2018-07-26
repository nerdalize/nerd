package svc

import (
    "context"
    "fmt"

    "github.com/nerdalize/nerd/pkg/kubevisor"

    "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/api/resource"
)

//CreateFileSystemInput is the input to CreateFileSystem
type CreateFileSystemInput struct {
    Name     string `validate:"printascii"`
    Capacity string `validate:"required"`
}

//CreateFileSystemOutput is the output to CreateFileSystem
type CreateFileSystemOutput struct {
    Name string
}

//CreateFileSystem will create a pvc on kubernetes
func (k *Kube) CreateFileSystem(ctx context.Context, in *CreateFileSystemInput) (out *CreateFileSystemOutput, err error) {
    if err = k.checkInput(ctx, in); err != nil {
        return nil, err
    }

    quantity, err := resource.ParseQuantity(in.Capacity)
    if err != nil {
        return nil, fmt.Errorf("invalid capacity, expected Kubernetes style resource requirement format, got: %v", in.Capacity)
    }

    storageClassName := "nerdalize"
    volumeMode := v1.PersistentVolumeFilesystem

    pvc := &v1.PersistentVolumeClaim{
        ObjectMeta: metav1.ObjectMeta{
            Labels: map[string]string{"storage": "nerdalize"},
        },
        Spec: v1.PersistentVolumeClaimSpec{
            AccessModes: []v1.PersistentVolumeAccessMode{
                v1.ReadWriteMany,
            },
            Resources: v1.ResourceRequirements{
                Requests: v1.ResourceList{
                    v1.ResourceName(v1.ResourceStorage): quantity,
                },
            },
            StorageClassName: &storageClassName,
            VolumeMode: &volumeMode,
        },
    }

    err = k.visor.CreateResource(ctx, kubevisor.ResourceTypePersistentVolumeClaims, pvc, in.Name)
    if err != nil {
        return nil, err
    }

    return &CreateFileSystemOutput{
        Name: pvc.Name,
    }, nil
}