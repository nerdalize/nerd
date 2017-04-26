package v1batch

import (
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//CreateDataset creates a new dataset.
func (c *Client) CreateDataset(projectID string) (output *v1payload.DatasetCreateOutput, err error) {
	output = &v1payload.DatasetCreateOutput{}
	return output, c.doRequest(http.MethodPost, createPath(projectID, datasetEndpoint), nil, output)
}

//GetDataset gets a dataset by ID.
func (c *Client) GetDataset(projectID, id string) (output *v1payload.DatasetDescribeOutput, err error) {
	output = &v1payload.DatasetDescribeOutput{}
	return output, c.doRequest(http.MethodGet, createPath(projectID, datasetEndpoint, id), nil, output)
}
