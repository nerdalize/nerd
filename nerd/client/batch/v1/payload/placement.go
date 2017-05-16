package v1payload

//CreatePlacementInput is input for queue creation
type CreatePlacementInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	Host      string `json:"host" valid:"required"`
	Token     string `json:"token" valid:"required"`
	CAPem     string `json:"ca_pem"`
}

//CreatePlacementOutput is output for queue creation
type CreatePlacementOutput struct {
}

//DeletePlacementInput is input for queue creation
type DeletePlacementInput struct {
	ProjectID string `json:"project_id" valid:"required"`
}

//DeletePlacementOutput is output for queue creation
type DeletePlacementOutput struct{}
