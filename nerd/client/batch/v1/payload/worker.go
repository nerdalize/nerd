package v1payload

//StartWorkerInput is input for queue creation
type StartWorkerInput struct {
	ProjectID      string            `json:"project_id" valid:"required"`
	Image          string            `json:"image" valid:"required"`
	QueueID        string            `json:"queue_id" valid:"required"`
	InputDatasetID string            `json:"input_dataset_id"`
	Env            map[string]string `json:"env"`
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

//WorkerSummary is a smaller representation of a queue
type WorkerSummary struct {
	ProjectID string `json:"project_id"`
	WorkerID  string `json:"worker_id"`
}

//ListWorkersInput is input for queue creation
type ListWorkersInput struct {
	ProjectID string `json:"project_id" valid:"required"`
}

//ListWorkersOutput is output for queue creation
type ListWorkersOutput struct {
	Workers []*WorkerSummary
}
