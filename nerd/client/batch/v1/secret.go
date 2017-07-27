package v1batch

import (
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

// ClientSecretInterface is an interface so client secret calls can be mocked.
type ClientSecretInterface interface {
	CreateSecret(projectID, name, key, value string) (output *v1payload.CreateSecretOutput, err error)
	DeleteSecret(projectID, name string) (output *v1payload.DeleteSecretOutput, err error)
	DescribeSecret(projectID, name string) (output *v1payload.DescribeSecretOutput, err error)
	ListSecrets(projectID string) (output *v1payload.ListSecretsOutput, err error)
}

// CreateSecret will create secret
func (c *Client) CreateSecret(projectID, name, key, value string) (output *v1payload.CreateSecretOutput, err error) {
	output = &v1payload.CreateSecretOutput{}
	input := &v1payload.CreateSecretInput{
		ProjectID: projectID,
		Name:      name,
		Key:       key,
		Value:     value,
		Type:      v1payload.SecretTypeOpaque,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, secretsEndpoint), input, output)
}

// CreatePullSecret will create a pull secret
func (c *Client) CreatePullSecret(projectID, name, dockerUsername, dockerPassword string) (output *v1payload.CreateSecretOutput, err error) {
	output = &v1payload.CreateSecretOutput{}
	input := &v1payload.CreateSecretInput{
		ProjectID:      projectID,
		Name:           name,
		Type:           v1payload.SecretTypeRegistry,
		DockerUsername: dockerUsername,
		DockerPassword: dockerPassword,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, secretsEndpoint), input, output)
}

//DeleteSecret will delete a secret with the provided name
func (c *Client) DeleteSecret(projectID, name string) (output *v1payload.DeleteSecretOutput, err error) {
	output = &v1payload.DeleteSecretOutput{}
	input := &v1payload.DeleteSecretInput{
		ProjectID: projectID,
		Name:      name,
	}

	return output, c.doRequest(http.MethodDelete, createPath(projectID, secretsEndpoint, name), input, output)
}

// ListSecrets will return all secrets for a particular project
func (c *Client) ListSecrets(projectID string) (output *v1payload.ListSecretsOutput, err error) {
	output = &v1payload.ListSecretsOutput{}
	input := &v1payload.ListSecretsInput{
		ProjectID: projectID,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, secretsEndpoint), input, output)
}

// DescribeSecret returns detailed information of a secret
func (c *Client) DescribeSecret(projectID, name string) (output *v1payload.DescribeSecretOutput, err error) {
	output = &v1payload.DescribeSecretOutput{}
	input := &v1payload.DescribeSecretInput{
		ProjectID: projectID,
		Name:      name,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, secretsEndpoint, name), input, output)
}
