package v1payload

//StartWorkerInput is input for queue creation
type StartWorkerInput struct {
	ProjectID string `json:"project_id" valid:"required"`
}

//StartWorkerOutput is output for queue creation
type StartWorkerOutput struct {
	ProjectID string `json:"project_id" valid:"required"`
	WorkerID  string `json:"worker_id" valid:"required"`
}

//StopWorkerInput is input for queue creation
type StopWorkerInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	WorkerID  string `json:"worker_id" valid:"required"`
}

//StopWorkerOutput is output for queue creation
type StopWorkerOutput struct{}
