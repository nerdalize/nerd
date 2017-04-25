package v1batch

import (
	"net/http"

	v2payload "github.com/nerdalize/nerd/nerd/payload/v2"
)

//CreateDataset creates a new dataset.
func (c *Client) CreateDataset(projectID string) (output *v2payload.DatasetCreateOutput, err error) {
	output = &v2payload.DatasetCreateOutput{}
	return output, c.doRequest(http.MethodPost, createPath(projectID, datasetEndpoint), nil, output)
}

//GetDataset gets a dataset by ID.
func (c *Client) GetDataset(projectID, id string) (output *v2payload.DatasetDescribeOutput, err error) {
	output = &v2payload.DatasetDescribeOutput{}
	return output, c.doRequest(http.MethodGet, createPath(projectID, datasetEndpoint, id), nil, output)
}
