package payload

//An Run acts as an reference to a task instance
type Run struct {
	ProjectID string `json:"project_id"`
	QueueID   string `json:"queue_id"`
	TaskID    string `json:"task_id"`
	Token     string `json:"token"`
	Payload   string `json:"payload"`
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
