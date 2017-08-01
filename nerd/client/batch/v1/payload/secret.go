package v1payload

const (
	//SecretTypeRegistry is the type for registring images
	SecretTypeRegistry string = "registry"

	//SecretTypeOpaque is the type that allows secrets that are opaque
	SecretTypeOpaque string = "opaque"
)

// CreateSecretInput is the input for creating a secret
type CreateSecretInput struct {
	ProjectID      string `json:"project_id" valid:"required"`
	Name           string `json:"name" valid:"required"`
	Key            string `json:"key"`
	Value          string `json:"value"`
	DockerUsername string `json:"dockerUsername"`
	DockerPassword string `json:"dockerPassword"`
	Type           string `json:"type" valid:"required,in(opaque|registry)"`
}

// CreateSecretOutput is the output from creating a secret
type CreateSecretOutput struct {
	ProjectID string `json:"project_id" valid:"required"`
	Name      string `json:"name" valid:"required"`
	Type      string `json:"type" valid:"required,in(opaque|registry)"`
}

// DescribeSecretInput is the input for describing a secret
type DescribeSecretInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	Name      string `json:"name" valid:"required"`
}

// DescribeSecretOutput is the output from describing a secret
type DescribeSecretOutput struct {
	ProjectID      string `json:"project_id" valid:"required"`
	Name           string `json:"name" valid:"required"`
	Key            string `json:"key"`
	Value          string `json:"value"`
	DockerUsername string `json:"dockerUsername"`
	DockerPassword string `json:"dockerPassword"`
	Type           string `json:"type" valid:"required,in(opaque|registry)"`
}

// DeleteSecretInput is the input for deleting a secret
type DeleteSecretInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	Name      string `json:"name" valid:"required"`
}

// DeleteSecretOutput is the output from deleting a secret
type DeleteSecretOutput struct {
}

//SecretSummary is summary of a secret
type SecretSummary struct {
	ProjectID string `json:"project_id" valid:"required"`
	Name      string `json:"name" valid:"required"`
	Type      string `json:"type" valid:"required,in(opaque|registry)"`
}

// ListSecretsInput is the input for listing secrets
type ListSecretsInput struct {
	ProjectID string `json:"project_id" valid:"required"`
}

// ListSecretsOutput is the output from listing secrets
type ListSecretsOutput struct {
	ProjectID string           `json:"project_id" valid:"required"`
	Secrets   []*SecretSummary `json:"secrets" valid:"required"`
}
