package v1batch

import (
	"net/http"
	"path"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//ClientCapacityInterface is an interface so client queue calls can be mocked.
type ClientCapacityInterface interface {
	ClaimCapacity(clusterID string) (output *v1payload.ClaimCapacityOutput, err error)
	ReleaseCapacity(clusterID, capacityID string) (output *v1payload.ReleaseCapacityOutput, err error)
}

//ClaimCapacity will create capacity
func (c *Client) ClaimCapacity(clusterID string) (output *v1payload.ClaimCapacityOutput, err error) {
	output = &v1payload.ClaimCapacityOutput{}
	input := &v1payload.ClaimCapacityInput{
		ClusterID: clusterID,
	}

	return output, c.doRequest(http.MethodPost, path.Join("clusters", clusterID, "capacity"), input, output)
}

//ReleaseCapacity will release compute capacity
func (c *Client) ReleaseCapacity(clusterID, capacityID string) (output *v1payload.ReleaseCapacityOutput, err error) {
	output = &v1payload.ReleaseCapacityOutput{}
	input := &v1payload.ReleaseCapacityInput{
		CapacityID: capacityID,
		ClusterID:  clusterID,
	}

	return output, c.doRequest(http.MethodDelete, path.Join("clusters", clusterID, "capacity", capacityID), input, output)
}
