package v1batch

import (
	"net/http"
	"strconv"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientRunInterface is an interface so client task calls can be mocked.
type ClientRunInterface interface {
	SendRunHeartbeat(projectID, workloadID string, taskID int64, runToken string) (output *v1payload.SendRunHeartbeatOutput, err error)
	SendRunSuccess(projectID, workloadID string, taskID int64, runToken, result string) (output *v1payload.SendRunSuccessOutput, err error)
	SendRunFailure(projectID, workloadID string, taskID int64, runToken, errCode, errMessage string) (output *v1payload.SendRunFailureOutput, err error)
}

//SendRunHeartbeat will send a heartbeat for a task run
func (c *Client) SendRunHeartbeat(projectID, workloadID string, taskID int64, runToken string) (output *v1payload.SendRunHeartbeatOutput, err error) {
	output = &v1payload.SendRunHeartbeatOutput{}
	input := &v1payload.SendRunHeartbeatInput{
		TaskID:     taskID,
		ProjectID:  projectID,
		WorkloadID: workloadID,
		RunToken:   runToken,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, workloadsEndpoint, workloadID, "tasks", strconv.FormatInt(taskID, 10), "heartbeats"), input, output)
}

//SendRunSuccess will send a successfully run for a task
func (c *Client) SendRunSuccess(projectID, workloadID string, taskID int64, runToken, result string) (output *v1payload.SendRunSuccessOutput, err error) {
	output = &v1payload.SendRunSuccessOutput{}
	input := &v1payload.SendRunSuccessInput{
		TaskID:     taskID,
		ProjectID:  projectID,
		WorkloadID: workloadID,
		RunToken:   runToken,
		Result:     result,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, workloadsEndpoint, workloadID, "tasks", strconv.FormatInt(taskID, 10), "success"), input, output)
}

//SendRunFailure will send a failure for a run
func (c *Client) SendRunFailure(projectID, workloadID string, taskID int64, runToken, errCode, errMessage string) (output *v1payload.SendRunFailureOutput, err error) {
	output = &v1payload.SendRunFailureOutput{}
	input := &v1payload.SendRunFailureInput{
		TaskID:       taskID,
		ProjectID:    projectID,
		WorkloadID:   workloadID,
		RunToken:     runToken,
		ErrorCode:    errCode,
		ErrorMessage: errMessage,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, workloadsEndpoint, workloadID, "tasks", strconv.FormatInt(taskID, 10), "failure"), input, output)
}
