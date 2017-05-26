package v1payload

//SecretSummary is summary of a secret
type SecretSummary struct {
	Name  string `json:"name" valid:"required"`
	Key   string `json:"key" valid:"required"`
	Value string `json:"value" valid:"required"`
}

// CreateSecretInput is the input for creating a secret
type CreateSecretInput struct {
	Name  string `json:"name" valid:"required"`
	Key   string `json:"key" valid:"required"`
	Value string `json:"value" valid:"required"`
}

// CreateSecretOutput is the output from creating a secret
type CreateSecretOutput struct {
	ProjectID string `json:"project_id" valid:"required"`
	Name      string `json:"name" valid:"required"`
	Key       string `json:"key" valid:"required"`
	Value     string `json:"value" valid:"required"`
}

// DeleteSecretInput is the input for deleting a secret
type DeleteSecretInput struct {
	Name string `json:"name" valid:"required"`
}

// DeleteSecretOutput is the output from deleting a secret
type DeleteSecretOutput struct {
}

// GetSecretInput is the input for getting a secret
type GetSecretInput struct {
	Name string `json:"name" valid:"required"`
}

// GetSecretOutput is the output from getting a secret
type GetSecretOutput struct {
	ProjectID string `json:"project_id" valid:"required"`
	Name      string `json:"name" valid:"required"`
	Key       string `json:"key" valid:"required"`
	Value     string `json:"value" valid:"required"`
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
