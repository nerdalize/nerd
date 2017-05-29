package v1batch

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientTaskInterface is an interface so client task calls can be mocked.
type ClientTaskInterface interface {
	StartTask(projectID, workloadID string, cmd []string, env map[string]string, stdin []byte) (output *v1payload.StartTaskOutput, err error)
	StopTask(projectID, workloadID string, taskID int64) (output *v1payload.StopTaskOutput, err error)
	ListTasks(projectID, workloadID string) (output *v1payload.ListTasksOutput, err error)
	DescribeTask(projectID, workloadID string, taskID int64) (output *v1payload.DescribeTaskOutput, err error)
	ReceiveTaskRuns(projectID, workloadID string, timeout time.Duration, queueOps QueueOps) (output []*v1payload.Run, err error)
}

// QueueOps is an interface that includes queue operations.
type QueueOps interface {
	ReceiveMessages(queueURL string, maxNoOfMessages, waitTimeSeconds int64) (messages []interface{}, err error)
	UnmarshalMessage(message interface{}, v interface{}) error
	DeleteMessage(queueURL string, message interface{}) error
}

//DescribeTask will create an execute a new task
func (c *Client) DescribeTask(projectID, workloadID string, taskID int64) (output *v1payload.DescribeTaskOutput, err error) {
	output = &v1payload.DescribeTaskOutput{}
	input := &v1payload.DescribeTaskInput{
		ProjectID:  projectID,
		WorkloadID: workloadID,
		TaskID:     taskID,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, workloadsEndpoint, workloadID, "tasks", strconv.FormatInt(taskID, 10)), input, output)
}

//StartTask will create an execute a new task
func (c *Client) StartTask(projectID, workloadID string, cmd []string, env map[string]string, stdin []byte) (output *v1payload.StartTaskOutput, err error) {
	output = &v1payload.StartTaskOutput{}
	input := &v1payload.StartTaskInput{
		WorkloadID: workloadID,
		ProjectID:  projectID,
		Cmd:        cmd,
		Env:        env,
		Stdin:      stdin,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, workloadsEndpoint, workloadID, "tasks"), input, output)
}

//StopTask will create queue
func (c *Client) StopTask(projectID, workloadID string, taskID int64) (output *v1payload.StopTaskOutput, err error) {
	output = &v1payload.StopTaskOutput{}
	input := &v1payload.StopTaskInput{
		ProjectID:  projectID,
		WorkloadID: workloadID,
		TaskID:     taskID,
	}

	return output, c.doRequest(http.MethodDelete, createPath(projectID, workloadsEndpoint, workloadID, "tasks", strconv.FormatInt(taskID, 10)), input, output)
}

// ListTasks will return all tasks in a queue
func (c *Client) ListTasks(projectID, workloadID string) (output *v1payload.ListTasksOutput, err error) {
	output = &v1payload.ListTasksOutput{}
	input := &v1payload.ListTasksInput{
		ProjectID:  projectID,
		WorkloadID: workloadID,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, workloadsEndpoint, workloadID, "tasks"), input, output)
}

//PatchTask will patch a task
func (c *Client) PatchTask(projectID, workloadID string, taskID int64, outputDatasetID string) (output *v1payload.PatchTaskOutput, err error) {
	output = &v1payload.PatchTaskOutput{}
	input := &v1payload.PatchTaskInput{
		WorkloadID:      workloadID,
		ProjectID:       projectID,
		TaskID:          taskID,
		OutputDatasetID: outputDatasetID,
	}

	return output, c.doRequest(http.MethodPatch, createPath(projectID, workloadsEndpoint, workloadID, "tasks", strconv.FormatInt(taskID, 10)), input, output)
}

//ReceiveTaskRuns will long poll the aws sqs queue for the availability of new runs. It will receive and delete messages once decoded
func (c *Client) ReceiveTaskRuns(projectID, workloadID string, timeout time.Duration, queueOps QueueOps) (output []*v1payload.Run, err error) {
	workload, err := c.DescribeWorkload(projectID, workloadID)
	if err != nil {
		return nil, fmt.Errorf("failed to describe workload: %+v", err)
	}

	toCh := time.After(timeout)
	for {
		select {
		case <-toCh:
			return output, nil
		default:
		}

		out, err := queueOps.ReceiveMessages(workload.QueueURL, 1, 5)
		if err != nil {
			return nil, fmt.Errorf("failed to receive runs: %+v", err)
		}

		for _, msg := range out {
			r := &v1payload.Run{}
			err = queueOps.UnmarshalMessage(msg, r)
			if err != nil {
				return nil, fmt.Errorf("failed to decode message: %+v", err)
			}

			if err = queueOps.DeleteMessage(workload.QueueURL, msg); err != nil {
				return nil, fmt.Errorf("failed to receive runs: %+v", err)
			}

			hb, err := c.SendRunHeartbeat(r.ProjectID, r.WorkloadID, r.TaskID, r.Token)
			if err != nil || hb.HasExpired {
				continue //we will not consider this run at all, it must be expired
			}

			output = append(output, r)
			return output, nil
		}
	}
}
