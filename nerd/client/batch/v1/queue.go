package v1batch

import (
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientQueueInterface is an interface so client queue calls can be mocked.
type ClientQueueInterface interface {
	CreateQueue(projectID string) (output *v1payload.CreateQueueOutput, err error)
	DeleteQueue(projectID, queueID string) (output *v1payload.DeleteQueueOutput, err error)
	DescribeQueue(projectID, queueID string) (output *v1payload.DescribeQueueOutput, err error)
	ListQueues(projectID string) (output *v1payload.ListQueuesOutput, err error)
}

//CreateQueue will create queue
func (c *Client) CreateQueue(projectID string) (output *v1payload.CreateQueueOutput, err error) {
	output = &v1payload.CreateQueueOutput{}
	input := &v1payload.CreateQueueInput{
		ProjectID: projectID,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, queuesEndpoint), input, output)
}

//DeleteQueue will delete queue a queue with the provided id
func (c *Client) DeleteQueue(projectID, queueID string) (output *v1payload.DeleteQueueOutput, err error) {
	output = &v1payload.DeleteQueueOutput{}
	input := &v1payload.DeleteQueueInput{
		ProjectID: projectID,
		QueueID:   queueID,
	}

	return output, c.doRequest(http.MethodDelete, createPath(projectID, queuesEndpoint, queueID), input, output)
}

// ListQueues will return all tasks in a queue
func (c *Client) ListQueues(projectID string) (output *v1payload.ListQueuesOutput, err error) {
	output = &v1payload.ListQueuesOutput{}
	input := &v1payload.ListQueuesInput{
		ProjectID: projectID,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, queuesEndpoint), input, output)
}

//DescribeQueue returns detailed information of a queue
func (c *Client) DescribeQueue(projectID, queueID string) (output *v1payload.DescribeQueueOutput, err error) {
	output = &v1payload.DescribeQueueOutput{}
	input := &v1payload.DescribeQueueInput{
		ProjectID: projectID,
		QueueID:   queueID,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, queuesEndpoint, queueID), input, output)
}
