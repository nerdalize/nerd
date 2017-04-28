package v1batch

import (
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientQueueInterface is an interface so client token calls can be mocked.
type ClientTokenInterface interface {
	CreateToken(projectID string) (output *v1payload.CreateTokenOutput, err error)
}

//CreateToken will create queue
func (c *Client) CreateToken(projectID string) (output *v1payload.CreateTokenOutput, err error) {
	output = &v1payload.CreateTokenOutput{}
	input := &v1payload.CreateTokenInput{
		ProjectID: projectID,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, tokensEndpoint), input, output)
}
