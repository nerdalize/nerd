package v1batch

import (
	"net/http"
	"path"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientClusterInterface is an interface so client queue calls can be mocked.
type ClientClusterInterface interface {
	RegisterCluster(host, token, capem string) (output *v1payload.RegisterClusterOutput, err error)
	DeregisterCluster(clusterID string) (output *v1payload.DeregisterClusterOutput, err error)
}

//RegisterCluster will create queue
func (c *Client) RegisterCluster(host, token, capem string) (output *v1payload.RegisterClusterOutput, err error) {
	output = &v1payload.RegisterClusterOutput{}
	input := &v1payload.RegisterClusterInput{
		Host:  host,
		Token: token,
		CAPem: capem,
	}

	return output, c.doRequest(http.MethodPost, path.Join("clusters"), input, output)
}

//DeregisterCluster will delete queue a queue with the provided id
func (c *Client) DeregisterCluster(clusterID string) (output *v1payload.DeregisterClusterOutput, err error) {
	output = &v1payload.DeregisterClusterOutput{}
	input := &v1payload.DeregisterClusterInput{
		ClusterID: clusterID,
	}

	return output, c.doRequest(http.MethodDelete, path.Join("clusters", clusterID), input, output)
}
