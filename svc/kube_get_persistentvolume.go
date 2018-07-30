package svc

import (
    "context"
    "strconv"

    "github.com/nerdalize/nerd/pkg/kubevisor"

    "k8s.io/api/core/v1"
)

//GetPersistentVolumeInput is the input to GetPersistentVolume
type GetPersistentVolumeInput struct {
    Name string `validate:"printascii"`
}

//GetPersistentVolumeOutput is the output to GetPersistentVolume
type GetPersistentVolumeOutput struct {
    Name string
    
    WebDAVHost string
    WebDAVPort int
    WebDAVPath string
}

//GetPersistentVolume will retrieve a persistent volume from kubernetes
func (k *Kube) GetPersistentVolume(ctx context.Context, in *GetPersistentVolumeInput) (out *GetPersistentVolumeOutput, err error) {
    if err = k.checkInput(ctx, in); err != nil {
        return nil, err
    }

    // TODO: Actually add prefix to PV name
    pv := &v1.PersistentVolume{}
    err = k.visor.GetClusterResource(ctx, kubevisor.ResourceTypePersistentVolumes, pv, in.Name)
    if err != nil {
        return nil, err
    }

    return GetPersistentVolumeOutputFromSpec(pv), nil
}

//GetPersistentVolumeOutputFromSpec allows easy output creation from dataset
func GetPersistentVolumeOutputFromSpec(pv *v1.PersistentVolume) *GetPersistentVolumeOutput {
    intPort, _ := strconv.Atoi(pv.ObjectMeta.Annotations["webdavPort"])

    return &GetPersistentVolumeOutput{
        Name:       pv.Name,
        WebDAVHost: pv.ObjectMeta.Annotations["webdavHost"],
        WebDAVPort: intPort,
        WebDAVPath: pv.ObjectMeta.Annotations["webdavPath"],
    }
}
