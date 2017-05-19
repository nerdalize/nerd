package v1payload

//CreateWorkerInput is used as input to worker creation
type CreateWorkerInput struct {
	ProjectID string            `json:"project_id" valid:"required"`
	Env       map[string]string `json:"env"`
}

//CreateWorkerOutput is returned from creating a worker
type CreateWorkerOutput struct {
	WorkerSummary
}

//WorkerSummary is a small version of
type WorkerSummary struct {
	ProjectID string `json:"project_id"`
}
