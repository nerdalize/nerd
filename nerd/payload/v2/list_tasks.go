package payload

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
