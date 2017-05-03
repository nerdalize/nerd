package v1batch

import (
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientUploadInterface is an interface so client task calls can be mocked.
type ClientUploadInterface interface {
	SendUploadHeartbeat(projectID, datasetID string, runToken string) (output *v1payload.SendUploadHeartbeatOutput, err error)
	SendUploadSuccess(projectID, datasetID string, runToken, result string) (output *v1payload.SendUploadSuccessOutput, err error)
}

//SendUploadHeartbeat will send a heartbeat for a task run
func (c *Client) SendUploadHeartbeat(projectID, datasetID string, runToken string) (output *v1payload.SendUploadHeartbeatOutput, err error) {
	output = &v1payload.SendUploadHeartbeatOutput{}
	input := &v1payload.SendUploadHeartbeatInput{
		ProjectID: projectID,
		DatasetID: datasetID,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, datasetEndpoint, datasetID, "heartbeats"), input, output)
}

//SendUploadSuccess will send a successfully run for a task
func (c *Client) SendUploadSuccess(projectID, datasetID string, runToken, result string) (output *v1payload.SendUploadSuccessOutput, err error) {
	output = &v1payload.SendUploadSuccessOutput{}
	input := &v1payload.SendUploadSuccessInput{
		ProjectID: projectID,
		DatasetID: datasetID,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, datasetEndpoint, datasetID, "success"), input, output)
}
