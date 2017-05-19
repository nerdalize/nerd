package v1batch

import (
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientPlacementInterface is an interface for placement of project
type ClientPlacementInterface interface {
	CreatePlacement(projectID, host, token, capem string) (output *v1payload.CreatePlacementOutput, err error)
	DeletePlacement(projectID string) (output *v1payload.DeletePlacementOutput, err error)
}

//CreatePlacement will create queue
func (c *Client) CreatePlacement(projectID, host, token, capem string) (output *v1payload.CreatePlacementOutput, err error) {
	output = &v1payload.CreatePlacementOutput{}
	input := &v1payload.CreatePlacementInput{
		ProjectID: projectID,
		Host:      host,
		Token:     token,
		CAPem:     capem,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, placementsEndpoint), input, output)
}

//DeletePlacement will delete queue a queue with the provided id
func (c *Client) DeletePlacement(projectID string) (output *v1payload.DeletePlacementOutput, err error) {
	output = &v1payload.DeletePlacementOutput{}
	input := &v1payload.DeletePlacementInput{
		ProjectID: projectID,
	}

	return output, c.doRequest(http.MethodDelete, createPath(projectID, placementsEndpoint), input, output)
}
