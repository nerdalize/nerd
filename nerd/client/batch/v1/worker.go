package v1batch

import (
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientWorkerInterface is an interface so client workload calls can be mocked.
type ClientWorkerInterface interface {
	WorkerLogs(projectID, workloadID, workerID string) (output *v1payload.WorkerLogsOutput, err error)
}

// WorkerLogs will return all workloads
func (c *Client) WorkerLogs(projectID, workloadID, workerID string) (output *v1payload.WorkerLogsOutput, err error) {
	output = &v1payload.WorkerLogsOutput{}
	input := &v1payload.WorkerLogsInput{
		ProjectID:  projectID,
		WorkloadID: workloadID,
		WorkerID:   workerID,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, workloadsEndpoint, workloadID, "workers", workerID, "logs"), input, output)
}
