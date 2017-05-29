package v1batch

import (
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientWorkloadInterface is an interface so client workload calls can be mocked.
type ClientWorkloadInterface interface {
	StartWorkload(projectID, image, inputDatasetID string, env map[string]string) (output *v1payload.StartWorkloadOutput, err error)
	StopWorkload(projectID, workloadID string) (output *v1payload.StopWorkloadOutput, err error)
	ListWorkloads(projectID string) (output *v1payload.ListWorkloadsOutput, err error)
	DescribeWorkload(projectID, workloadID string) (output *v1payload.DescribeWorkloadOutput, err error)
}

//StartWorkload will start a workload
func (c *Client) StartWorkload(projectID, image, inputDatasetID string, env map[string]string) (output *v1payload.StartWorkloadOutput, err error) {
	output = &v1payload.StartWorkloadOutput{}
	input := &v1payload.StartWorkloadInput{
		ProjectID:      projectID,
		Image:          image,
		InputDatasetID: inputDatasetID,
		Env:            env,
	}

	return output, c.doRequest(http.MethodPost, createPath(projectID, workloadsEndpoint), input, output)
}

//StopWorkload will stop a workload
func (c *Client) StopWorkload(projectID, workloadID string) (output *v1payload.StopWorkloadOutput, err error) {
	output = &v1payload.StopWorkloadOutput{}
	input := &v1payload.StopWorkloadInput{
		ProjectID:  projectID,
		WorkloadID: workloadID,
	}

	return output, c.doRequest(http.MethodDelete, createPath(projectID, workloadsEndpoint, workloadID), input, output)
}

// ListWorkloads will return all workloads
func (c *Client) ListWorkloads(projectID string) (output *v1payload.ListWorkloadsOutput, err error) {
	output = &v1payload.ListWorkloadsOutput{}
	input := &v1payload.ListWorkloadsInput{
		ProjectID: projectID,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, workloadsEndpoint), input, output)
}

//DescribeWorkload returns detailed information of a workload
func (c *Client) DescribeWorkload(projectID, workloadID string) (output *v1payload.DescribeWorkloadOutput, err error) {
	output = &v1payload.DescribeWorkloadOutput{}
	input := &v1payload.DescribeWorkloadInput{
		ProjectID:  projectID,
		WorkloadID: workloadID,
	}

	return output, c.doRequest(http.MethodGet, createPath(projectID, workloadsEndpoint, workloadID), input, output)
}
