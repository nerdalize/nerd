package v2client

import (
	"net/http"

	v2payload "github.com/nerdalize/nerd/nerd/payload/v2"
)

//CreateQueue will create queue
func (c *Nerd) CreateQueue(projectID string) (output *v2payload.CreateQueueOutput, err error) {
	output = &v2payload.CreateQueueOutput{}
	input := &v2payload.CreateQueueInput{
		ProjectID: projectID,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, queuesEndpoint), input, output)
}

//DeleteQueue will delete queue a queue with the provided id
func (c *Nerd) DeleteQueue(projectID, queueID string) (output *v2payload.DeleteQueueOutput, err error) {
	output = &v2payload.DeleteQueueOutput{}
	input := &v2payload.DeleteQueueInput{
		ProjectID: projectID,
		QueueID:   queueID,
	}

	return output, c.doRequest(http.MethodDelete, createPath(projectID, queuesEndpoint, queueID), input, output)
}

//DescribeQueue returns detailed information of a queue
func (c *Nerd) DescribeQueue(projectID, queueID string) (output *v2payload.DescribeQueueOutput, err error) {
	output = &v2payload.DescribeQueueOutput{}
	input := &v2payload.DescribeQueueInput{
		ProjectID: projectID,
		QueueID:   queueID,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, queuesEndpoint, queueID), input, output)
}
