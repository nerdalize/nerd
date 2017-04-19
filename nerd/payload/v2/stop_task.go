package payload

//StopTaskInput is input for queue creation
type StopTaskInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	QueueID   string `json:"queue_id" valid:"required"`
	TaskID    string `json:"task_id" valid:"required"`
}

//StopTaskOutput is output for queue creation
type StopTaskOutput struct{}
