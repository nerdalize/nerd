package v1batch

import (
	"fmt"
	"net/http"
	"time"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientTaskInterface is an interface so client task calls can be mocked.
type ClientTaskInterface interface {
	StartTask(projectID, queueID, payload string) (output *v1payload.StartTaskOutput, err error)
	StopTask(projectID, queueID, taskID string) (output *v1payload.StopTaskOutput, err error)
	ListTasks(projectID, queueID string) (output *v1payload.ListTasksOutput, err error)
	KeepTask(projectID, queueID, taskID, runToken string) (output *v1payload.KeepTaskOutput, err error)
	ReceiveTaskRuns(projectID, queueID string, timeout time.Duration, queueOps QueueOps) (output []*v1payload.Run, err error)
}

// QueueOps is an interface that includes queue operations.
type QueueOps interface {
	ReceiveMessages(queueURL string, maxNoOfMessages, waitTimeSeconds int64) (messages []interface{}, err error)
	UnmarshalMessage(message interface{}, v interface{}) error
	DeleteMessage(queueURL string, message interface{}) error
}

//StartTask will create an execute a new task
func (c *Client) StartTask(projectID, queueID, payload string) (output *v1payload.StartTaskOutput, err error) {
	output = &v1payload.StartTaskOutput{}
	input := &v1payload.StartTaskInput{
		QueueID:   queueID,
		ProjectID: projectID,
		Payload:   payload,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, queuesEndpoint, queueID, "tasks"), input, output)
}

//StopTask will create queue
func (c *Client) StopTask(projectID, queueID, taskID string) (output *v1payload.StopTaskOutput, err error) {
	output = &v1payload.StopTaskOutput{}
	input := &v1payload.StopTaskInput{
		ProjectID: projectID,
		QueueID:   queueID,
		TaskID:    taskID,
	}

	return output, c.doRequest(http.MethodDelete, createPath(projectID, queuesEndpoint, queueID, "tasks"), input, output)
}

// ListTasks will return all tasks in a queue
func (c *Client) ListTasks(projectID, queueID string) (output *v1payload.ListTasksOutput, err error) {
	output = &v1payload.ListTasksOutput{}
	input := &v1payload.ListTasksInput{
		ProjectID: projectID,
		QueueID:   queueID,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, queuesEndpoint, queueID, "tasks"), input, output)
}

//ReceiveTaskRuns will long poll the aws sqs queue for the availability of new runs. It will receive and delete messages once decoded
func (c *Client) ReceiveTaskRuns(projectID, queueID string, timeout time.Duration, queueOps QueueOps) (output []*v1payload.Run, err error) {
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
		out, err := queueOps.ReceiveMessages(queue.QueueURL, 1, 5)
		if err != nil {
			return nil, fmt.Errorf("failed to receive runs: %+v", err)
		}

		for _, msg := range out {
			r := &v1payload.Run{}
			err = queueOps.UnmarshalMessage(msg, r)
			if err != nil {
				return nil, fmt.Errorf("failed to decode message: %+v", err)
			}

			if err = queueOps.DeleteMessage(queue.QueueURL, msg); err != nil {
				return nil, fmt.Errorf("failed to receive runs: %+v", err)
			}

			_, err = c.SendRunHeartbeat(r.ProjectID, r.QueueID, r.TaskID, r.Token)
			if err != nil {
				continue //we will not consider this run at all, it must be expired
			}

			output = append(output, r)
			return output, nil
		}
	}
}
