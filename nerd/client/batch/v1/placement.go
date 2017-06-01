package v1batch

import (
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientPlacementInterface is an interface for placement of project
type ClientPlacementInterface interface {
	PlaceProject(projectID, host, token, capem, username, password string, insecure bool) (output *v1payload.PlaceProjectOutput, err error)
	ExpelProject(projectID string) (output *v1payload.ExpelProjectOutput, err error)
}

//PlaceProject will create queue
func (c *Client) PlaceProject(projectID, host, token, capem, username, password string, insecure bool) (output *v1payload.PlaceProjectOutput, err error) {
	output = &v1payload.PlaceProjectOutput{}
	input := &v1payload.PlaceProjectInput{
		ProjectID: projectID,
		Host:      host,
		Token:     token,
		CAPem:     capem,
		Username:  username,
		Password:  password,
		Insecure:  insecure,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, placementsEndpoint), input, output)
}

//ExpelProject will delete queue a queue with the provided id
func (c *Client) ExpelProject(projectID string) (output *v1payload.ExpelProjectOutput, err error) {
	output = &v1payload.ExpelProjectOutput{}
	input := &v1payload.ExpelProjectInput{
		ProjectID: projectID,
	}

	return output, c.doRequest(http.MethodDelete, createPath(projectID, placementsEndpoint), input, output)
}
