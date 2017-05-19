package v1batch

import (
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientWorkerInterface is an interface so client worker calls can be mocked.
type ClientWorkerInterface interface {
	CreateWorker(projectID string) (output *v1payload.CreateWorkerOutput, err error)
}

//CreateWorker creates a new dataset.
func (c *Client) CreateWorker(projectID string, input *v1payload.CreateWorkerInput) (output *v1payload.CreateWorkerOutput, err error) {
	output = &v1payload.CreateWorkerOutput{}
	return output, c.doRequest(http.MethodPost, createPath(projectID, workersEndpoint), input, output)
}
