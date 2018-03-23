package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"
	"k8s.io/api/core/v1"
)

// UpdateSecretInput is the input for UpdateSecret
type UpdateSecretInput struct {
	Name     string `validate:"printascii"`
	Username string
	Password string
}

// UpdateSecretOutput is the output for UpdateSecret
type UpdateSecretOutput struct {
	Name string
}

// UpdateSecret will update a secret resource.
// Fields that can be updated: name, input, output and size. Input and output are the jobs the secret is used for or coming from.
func (k *Kube) UpdateSecret(ctx context.Context, in *UpdateSecretInput) (out *UpdateSecretOutput, err error) {
	secret := &v1.Secret{}
	err = k.visor.GetResource(ctx, kubevisor.ResourceTypeSecrets, secret, in.Name)
	if err != nil {
		return nil, err
	}

	secret.Data[v1.DockerConfigJsonKey], err = transformCredentials(in.Username, in.Password, secret.Labels["registry"])
	if err != nil {
		return nil, err
	}
	err = k.visor.UpdateResource(ctx, kubevisor.ResourceTypeSecrets, secret, in.Name)
	if err != nil {
		return nil, err
	}
	return &UpdateSecretOutput{
		Name: secret.Name,
	}, nil
}
