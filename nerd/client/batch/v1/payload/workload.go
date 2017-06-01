package v1payload

//WorkloadSummary is a smaller representation of a workload
type WorkloadSummary struct {
	ProjectID      string           `json:"project_id"`
	WorkloadID     string           `json:"workload_id"`
	QueueURL       string           `json:"queue_url"`
	Image          string           `json:"image"`
	NrOfWorkers    int              `json:"nr_of_workers"`
	InputDatasetID string           `json:"input_dataset_id"`
	CreatedAt      int64            `json:"created_at"`
	Workers        []*WorkerSummary `json:"workers"`
}

//ListWorkloadsInput is input for workload listing
type ListWorkloadsInput struct {
	ProjectID string `json:"project_id" valid:"required"`
}

//ListWorkloadsOutput is output for workload listing
type ListWorkloadsOutput struct {
	Workloads []*WorkloadSummary
}

//DescribeWorkloadInput is input for getting workload information
type DescribeWorkloadInput struct {
	ProjectID  string `json:"project_id" valid:"required"`
	WorkloadID string `json:"workload_id"`
}

//TaskStatus represents the status of a task
type TaskStatus string

//DescribeWorkloadOutput is output for getting workload information
type DescribeWorkloadOutput struct {
	WorkloadSummary
	TaskCount  map[TaskStatus]int `json:"task_count"`
	Env        map[string]string  `json:"env"`
	PullSecret string             `json:"pull_secret"`
}

//CreateWorkloadInput is input for workload creation
type CreateWorkloadInput struct {
	ProjectID      string            `json:"project_id" valid:"required"`
	Image          string            `json:"image" valid:"required"`
	NrOfWorkers    int               `json:"nr_of_workers" valid:"required"`
	InputDatasetID string            `json:"input_dataset_id"`
	UseCuteur      bool              `json:"use_cuteur"`
	Env            map[string]string `json:"env"`
	PullSecret     string            `json:"pull_secret"`
}

//CreateWorkloadOutput is output for workload creation
type CreateWorkloadOutput struct {
	WorkloadSummary
}

//StopWorkloadInput is input for workload deletion
type StopWorkloadInput struct {
	ProjectID  string `json:"project_id" valid:"required"`
	WorkloadID string `json:"workload_id" valid:"required"`
}

//StopWorkloadOutput is output for workload deletion
type StopWorkloadOutput struct{}
