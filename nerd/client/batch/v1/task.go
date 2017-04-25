package v1batch

import (
	"fmt"
	"net/http"
	"time"

	v2payload "github.com/nerdalize/nerd/nerd/payload/v2"
)

//StartTask will create an execute a new task
func (c *Client) StartTask(projectID, queueID, payload string) (output *v2payload.StartTaskOutput, err error) {
	output = &v2payload.StartTaskOutput{}
	input := &v2payload.StartTaskInput{
		QueueID:   queueID,
		ProjectID: projectID,
		Payload:   payload,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, queuesEndpoint, queueID, "tasks"), input, output)
}

//StopTask will create queue
func (c *Client) StopTask(projectID, queueID, taskID string) (output *v2payload.StopTaskOutput, err error) {
	output = &v2payload.StopTaskOutput{}
	input := &v2payload.StopTaskInput{
		ProjectID: projectID,
		QueueID:   queueID,
		TaskID:    taskID,
	}

	return output, c.doRequest(http.MethodDelete, createPath(projectID, queuesEndpoint, queueID, "tasks"), input, output)
}

// ListTasks will return all tasks in a queue
func (c *Client) ListTasks(projectID, queueID string) (output *v2payload.ListTasksOutput, err error) {
	output = &v2payload.ListTasksOutput{}
	input := &v2payload.ListTasksInput{
		ProjectID: projectID,
		QueueID:   queueID,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, queuesEndpoint, queueID, "tasks"), input, output)
}

// KeepTask will send a heartbeat
func (c *Client) KeepTask(projectID, queueID, taskID, runToken string) (output *v2payload.KeepTaskOutput, err error) {
	output = &v2payload.KeepTaskOutput{}
	input := &v2payload.KeepTaskInput{
		TaskID:    taskID,
		ProjectID: projectID,
		QueueID:   queueID,
		RunToken:  runToken,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, queuesEndpoint, queueID, "tasks", taskID, "heartbeats"), input, output)
}

//ReceiveTaskRuns will long poll the aws sqs queue for the availability of new runs. It will receive and delete messages once decoded
func (c *Client) ReceiveTaskRuns(projectID, queueID string, timeout time.Duration) (output []*v2payload.Run, err error) {
	queue, err := c.DescribeQueue(projectID, queueID)
	if err != nil {
		return nil, fmt.Errorf("failed to describe queue: %+v", err)
	}

	toCh := time.After(timeout)
	for {
		select {
		case <-toCh:
			return output, nil
		default:
		}

		// var out *sqs.ReceiveMessageOutput
		out, err := c.QueueOps.ReceiveMessages(queue.QueueURL, 1, 5)
		if err != nil {
			return nil, fmt.Errorf("failed to receive runs: %+v", err)
		}

		for _, msg := range out {
			r := &v2payload.Run{}
			err = c.QueueOps.UnmarshalMessage(msg, r)
			if err != nil {
				return nil, fmt.Errorf("failed to decode message: %+v", err)
			}

			if err = c.QueueOps.DeleteMessage(queue.QueueURL, msg); err != nil {
				return nil, fmt.Errorf("failed to receive runs: %+v", err)
			}

			_, err = c.KeepTask(r.ProjectID, r.QueueID, r.TaskID, r.Token)
			if err != nil {
				continue //we will not consider this run at all, it must be expired
			}

			output = append(output, r)
			return output, nil
		}
	}
}
