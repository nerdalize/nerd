package payload

type SecretSummary struct {
	Name  string `json:"name" valid:"required"`
	Key   string `json:"key" valid:"required"`
	Value string `json:"value" valid:"required"`
}

type CreateSecretInput struct {
	Name  string `json:"name" valid:"required"`
	Key   string `json:"key" valid:"required"`
	Value string `json:"value" valid:"required"`
}

type CreateSecretOutput struct {
	ProjectID string `json:"project_id" valid:"required"`
	Name      string `json:"name" valid:"required"`
	Key       string `json:"key" valid:"required"`
	Value     string `json:"value" valid:"required"`
}

type DeleteSecretInput struct {
	Name string `json:"name" valid:"required"`
}

type DeleteSecretOutput struct {
}

type GetSecretInput struct {
	Name string `json:"name" valid:"required"`
}

type GetSecretOutput struct {
	ProjectID string `json:"project_id" valid:"required"`
	Name      string `json:"name" valid:"required"`
	Key       string `json:"key" valid:"required"`
	Value     string `json:"value" valid:"required"`
}

type ListSecretsInput struct {
	ProjectID string `json:"project_id" valid:"required"`
}

type ListSecretsOutput struct {
	ProjectID string           `json:"project_id" valid:"required"`
	Secrets   []*SecretSummary `json:"secrets" valid:"required"`
}
