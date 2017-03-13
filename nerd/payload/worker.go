package payload

//WorkerCreateInput is used as input to worker creation
type WorkerCreateInput struct{}

//WorkerCreateOutput is returned from creating a worker
type WorkerCreateOutput struct {
	Worker
}

//WorkerDescribeOutput is returned from a specific worker
type WorkerDescribeOutput struct {
	Worker
}

//WorkerListOutput is returned from the worker listing
type WorkerListOutput struct {
	Workers []*Worker `json:"workers"`
}

//Worker is a worker in the list output
type Worker struct {
	ProjectID    string `json:"project_id"`
	WorkerID     string `json:"worker_id"`
	QueueURL     string `json:"queue_url"`
	LogGroupName string `json:"log_group_name"`
}
