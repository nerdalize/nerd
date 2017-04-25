package v2client

import (
	"net/http"

	v2payload "github.com/nerdalize/nerd/nerd/payload/v2"
)

//CreateToken will create queue
func (c *Nerd) CreateToken(projectID string) (output *v2payload.CreateTokenOutput, err error) {
	output = &v2payload.CreateTokenOutput{}
	input := &v2payload.CreateTokenInput{
		ProjectID: projectID,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, tokensEndpoint), input, output)
}
