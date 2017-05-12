package v1payload

//RegisterClusterInput is input for queue creation
type RegisterClusterInput struct {
	Host  string `json:"host" valid:"required"`
	Token string `json:"token" valid:"required"`
	CAPem string `json:"ca_pem" valid:"required"`
}

//RegisterClusterOutput is output for queue creation
type RegisterClusterOutput struct {
	ClusterID string `json:"cluster_id"`
}

//DeregisterClusterInput is input for queue creation
type DeregisterClusterInput struct {
	ClusterID string `json:"cluster_id" valid:"required"`
}

//DeregisterClusterOutput is output for queue creation
type DeregisterClusterOutput struct {
}
