package v1payload

//SendUploadHeartbeatInput is input for dataset creation
type SendUploadHeartbeatInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	DatasetID string `json:"dataset_id" valid:"required"`
}

//SendUploadHeartbeatOutput is output for dataset creation
type SendUploadHeartbeatOutput struct {
	HasExpired bool `json:"has_expired"`
}

//SendUploadSuccessInput is input for marking a run as failed
type SendUploadSuccessInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	DatasetID string `json:"dataset_id" valid:"required"`
}

//SendUploadSuccessOutput is output from marking a run as failed
type SendUploadSuccessOutput struct{}
