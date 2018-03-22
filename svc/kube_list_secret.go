package svc

import (
	"context"
	"path"
	"time"

	"github.com/nerdalize/nerd/pkg/kubevisor"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

//SecretDetails tells us more about the secret by looking at underlying resources
type SecretDetails struct {
	CreatedAt time.Time
	Size      int
	Type      string
	Image     string
}

//ListSecretItem is a secret listing item
type ListSecretItem struct {
	Name    string
	Details SecretDetails
}

//ListSecretsInput is the input to ListSecrets
type ListSecretsInput struct {
	Labels []string
}

//ListSecretsOutput is the output to ListSecrets
type ListSecretsOutput struct {
	Items []*ListSecretItem
}

//ListSecrets will create a secret on kubernetes
func (k *Kube) ListSecrets(ctx context.Context, in *ListSecretsInput) (out *ListSecretsOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	//Step 0: Get all the secrets under nerd-app=cli
	secrets := &secrets{}
	err = k.visor.ListResources(ctx, kubevisor.ResourceTypeSecrets, secrets, in.Labels, nil)
	if err != nil {
		return nil, err
	}

	//Step 1: Analyse secret structure and formulate our output items
	out = &ListSecretsOutput{}
	mapping := map[types.UID]*ListSecretItem{}
	for _, secret := range secrets.Items {
		if secret.Labels["registry"] == "index.docker.io" {
			secret.Labels["registry"] = ""
		}
		item := &ListSecretItem{
			Name: secret.GetName(),
			Details: SecretDetails{
				Type:      string(secret.Type),
				Size:      secret.Size(),
				CreatedAt: secret.CreationTimestamp.Local(),
				Image:     path.Join(secret.Labels["registry"], secret.Labels["project"], secret.Labels["image"]),
			},
		}

		mapping[secret.UID] = item
		out.Items = append(out.Items, item)
	}

	return out, nil
}

//secrets implements the list transformer interface to allow the kubevisor to manage names for us
type secrets struct{ *v1.SecretList }

func (secrets *secrets) Transform(fn func(in kubevisor.ManagedNames) (out kubevisor.ManagedNames)) {
	for i, d1 := range secrets.SecretList.Items {
		secrets.Items[i] = *(fn(&d1).(*v1.Secret))
	}
}

func (secrets *secrets) Len() int {
	return len(secrets.SecretList.Items)
}
