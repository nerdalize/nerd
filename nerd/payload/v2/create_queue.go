package payload

//CreateQueueInput is input for queue creation
type CreateQueueInput struct {
	ProjectID string `json:"project_id" valid:"required"`
}

//CreateQueueOutput is output for queue creation
type CreateQueueOutput struct {
	QueueID  string `json:"queue_id"`
	QueueURL string `json:"queue_url"`
}
