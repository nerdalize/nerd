package v1payload

//PlaceProjectInput is input for placement creation
type PlaceProjectInput struct {
	ProjectID    string `json:"project_id" valid:"required"`
	Host         string `json:"host" valid:"required"`
	Token        string `json:"token"`
	CAPem        string `json:"ca_pem"`
	Password     string `json:"password"`
	Username     string `json:"username"`
	Insecure     bool   `json:"insecure"`
	ComputeUnits string `json:"compute_units"`
}

//PlaceProjectOutput is output for placement creation
type PlaceProjectOutput struct {
}

//ExpelProjectInput is input for placement creation
type ExpelProjectInput struct {
	ProjectID string `json:"project_id" valid:"required"`
}

//ExpelProjectOutput is output for placement creation
type ExpelProjectOutput struct{}
