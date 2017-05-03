package v1batch

import (
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientDatasetInterface is an interface so client dataset calls can be mocked.
type ClientDatasetInterface interface {
	CreateDataset(projectID string) (output *v1payload.CreateDatasetOutput, err error)
	ListDatasets(projectID, tag string) (output *v1payload.ListDatasetsOutput, err error)
	DescribeDataset(projectID, id string) (output *v1payload.DescribeDatasetOutput, err error)
}

//CreateDataset creates a new dataset.
func (c *Client) CreateDataset(projectID string) (output *v1payload.CreateDatasetOutput, err error) {
	input := &v1payload.CreateDatasetInput{}
	output = &v1payload.CreateDatasetOutput{}
	return output, c.doRequest(http.MethodPost, createPath(projectID, datasetEndpoint), input, output)
}

//DescribeDataset gets a dataset by ID.
func (c *Client) DescribeDataset(projectID, id string) (output *v1payload.DescribeDatasetOutput, err error) {
	input := &v1payload.DescribeDatasetInput{}
	output = &v1payload.DescribeDatasetOutput{}
	return output, c.doRequest(http.MethodGet, createPath(projectID, datasetEndpoint, id), input, output)
}

//ListDatasets gets a dataset by ID.
func (c *Client) ListDatasets(projectID, tag string) (output *v1payload.ListDatasetsOutput, err error) {
	input := &v1payload.ListDatasetsInput{Tag: tag}
	output = &v1payload.ListDatasetsOutput{}
	return output, c.doRequest(http.MethodGet, createPath(projectID, datasetEndpoint), input, output)
}
