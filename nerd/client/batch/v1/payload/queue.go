package v1payload

//QueueSummary is a smaller representation of a queue
type QueueSummary struct {
	ProjectID string `json:"project_id"`
	QueueID   string `json:"queue_id"`
	QueueURL  string `json:"queue_url"`
}

//DescribeQueueInput is input for getting queue information
type DescribeQueueInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	QueueID   string `json:"queue_id"`
}

//DescribeQueueOutput is output for queue creation
type DescribeQueueOutput struct {
	QueueSummary
}

//CreateQueueInput is input for queue creation
type CreateQueueInput struct {
	ProjectID string `json:"project_id" valid:"required"`
}

//CreateQueueOutput is output for queue creation
type CreateQueueOutput struct {
	QueueSummary
}

//DeleteQueueInput is input for queue creation
type DeleteQueueInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	QueueID   string `json:"queue_id" valid:"required"`
}

//DeleteQueueOutput is output for queue creation
type DeleteQueueOutput struct{}
