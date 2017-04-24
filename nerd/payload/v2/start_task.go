package payload

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
