package svc

import (
	"context"
	"path"
	"time"

	"github.com/nerdalize/nerd/pkg/kubevisor"
	"k8s.io/api/core/v1"
)

//GetSecretInput is the input to GetSecret
type GetSecretInput struct {
	Name string `validate:"printascii"`
}

//GetSecretOutput is the output to GetSecret
type GetSecretOutput struct {
	Name      string
	Size      int
	Image     string
	CreatedAt time.Time
	Type      string
}

//GetSecret will retrieve the secret matching the provided name from kubernetes
func (k *Kube) GetSecret(ctx context.Context, in *GetSecretInput) (out *GetSecretOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	secret := &v1.Secret{}
	err = k.visor.GetResource(ctx, kubevisor.ResourceTypeSecrets, secret, in.Name)
	if err != nil {
		return nil, err
	}

	return &GetSecretOutput{
		Name:      secret.Name,
		Type:      string(secret.Type),
		Size:      secret.Size(),
		CreatedAt: secret.CreationTimestamp.Local(),
		Image:     path.Join(secret.Labels["registry"], secret.Labels["project"], secret.Labels["image"]),
	}, nil

}
