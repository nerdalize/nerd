package v1batch

import (
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientWorkerInterface is an interface for placement of project
type ClientWorkerInterface interface {
	StartWorker(projectID, image string, env map[string]string) (output *v1payload.StartWorkerOutput, err error)
	StopWorker(projectID, workerID string) (output *v1payload.StopWorkerOutput, err error)
	ListWorkers(projectID string) (output *v1payload.ListWorkersOutput, err error)
}

//StartWorker will create worker
func (c *Client) StartWorker(projectID, image string, env map[string]string) (output *v1payload.StartWorkerOutput, err error) {
	output = &v1payload.StartWorkerOutput{}
	input := &v1payload.StartWorkerInput{
		ProjectID: projectID,
		Image:     image,
		Env:       env,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, workersEndpoint), input, output)
}

//StopWorker will delete worker a worker with the provided id
func (c *Client) StopWorker(projectID, workerID string) (output *v1payload.StopWorkerOutput, err error) {
	output = &v1payload.StopWorkerOutput{}
	input := &v1payload.StopWorkerInput{
		ProjectID: projectID,
		WorkerID:  workerID,
	}

	return output, c.doRequest(http.MethodDelete, createPath(projectID, workersEndpoint, workerID), input, output)
}

// ListWorkers will return all tasks in a worker
func (c *Client) ListWorkers(projectID string) (output *v1payload.ListWorkersOutput, err error) {
	output = &v1payload.ListWorkersOutput{}
	input := &v1payload.ListWorkersInput{
		ProjectID: projectID,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, workersEndpoint), input, output)
}
