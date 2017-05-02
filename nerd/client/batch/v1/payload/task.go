package v1payload

//StopTaskInput is input for queue creation
type StopTaskInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	QueueID   string `json:"queue_id" valid:"required"`
	TaskID    string `json:"task_id" valid:"required"`
}

//StopTaskOutput is output for queue creation
type StopTaskOutput struct{}

//StartTaskInput is input for queue creation
type StartTaskInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	QueueID   string `json:"queue_id" valid:"required"`
	Payload   string `json:"payload" valid:"required"`
}

//StartTaskOutput is output for queue creation
type StartTaskOutput struct {
	TaskSummary
}

//ListTasksInput is input for queue creation
type ListTasksInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	QueueID   string `json:"queue_id" valid:"required"`
}

//TaskSummary is a small version of
type TaskSummary struct {
	TaskID  string `json:"task_id"`
	QueueID string `json:"queue_id"`
	Status  string `json:"status,omitempty"`
}

//ListTasksOutput is output for queue creation
type ListTasksOutput struct {
	Tasks []*TaskSummary
}

//KeepTaskInput is input for queue creation
type KeepTaskInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	QueueID   string `json:"queue_id" valid:"required"`
	TaskID    string `json:"task_id" valid:"required"`
	RunToken  string `json:"run_token" valid:"required"`
}

//KeepTaskOutput is output for queue creation
type KeepTaskOutput struct{}
