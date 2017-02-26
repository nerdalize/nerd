package payload

//WorkerCreateInput is used as input to worker creation
type WorkerCreateInput struct {
	ProjectID string `json:"project_id" valid:"min=1,required,uuid"`
}

//WorkerCreateOutput is returned from
type WorkerCreateOutput struct {
	Worker
}

//WorkerListOutput is returned from the worker listing
type WorkerListOutput struct {
	Workers []*Worker `json:"workers"`
}

//Worker is a worker in the list output
type Worker struct {
	ProjectID string `json:"project_id"`
	WorkerID  string `json:"worker_id"`
	QueueURL  string `json:"queue_url"`
}
