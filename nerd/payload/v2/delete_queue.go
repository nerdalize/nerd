package payload

//DeleteQueueInput is input for queue creation
type DeleteQueueInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	QueueID   string `json:"queue_id" valid:"required"`
}

//DeleteQueueOutput is output for queue creation
type DeleteQueueOutput struct{}
