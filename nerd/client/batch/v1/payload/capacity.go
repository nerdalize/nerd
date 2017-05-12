package v1payload

//ClaimCapacityInput is input for capacity creation
type ClaimCapacityInput struct {
	ClusterID string `json:"cluster_id" valid:"required"`
}

//ClaimCapacityOutput is output for capacity creation
type ClaimCapacityOutput struct {
}

//ReleaseCapacityInput is input for capacity creation
type ReleaseCapacityInput struct {
	ClusterID  string `json:"cluster_id" valid:"required"`
	CapacityID string `json:"capacity_id" valid:"required"`
}

//ReleaseCapacityOutput is output for capacity creation
type ReleaseCapacityOutput struct {
}
