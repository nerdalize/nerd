package v1payload

//StopTaskInput is input for queue creation
type StopTaskInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	QueueID   string `json:"queue_id" valid:"required"`
	TaskID    int64  `json:"task_id" valid:"required"`
}

//StopTaskOutput is output for queue creation
type StopTaskOutput struct{}

//StartTaskInput is input for queue creation
type StartTaskInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	QueueID   string `json:"queue_id" valid:"required"`

	Cmd   []string          `json:"cmd"`
	Env   map[string]string `json:"env"`
	Stdin []byte            `json:"stdin"`
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
	TaskID  int64  `json:"task_id"`
	QueueID string `json:"queue_id"`
	Status  string `json:"status,omitempty"`
}

//ListTasksOutput is output for queue creation
type ListTasksOutput struct {
	Tasks []*TaskSummary
}

//DescribeTaskInput is input for queue creation
type DescribeTaskInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	QueueID   string `json:"queue_id" valid:"required"`
	TaskID    int64  `json:"task_id" valid:"required"`
}

//DescribeTaskOutput is output for queue creation
type DescribeTaskOutput struct {
	TaskSummary
	ExecutionARN   string `json:"execution_arn"`
	NumDispatches  int64  `json:"num_dispatches"`
	Result         string `json:"result,omitempty"`
	LastErrCode    string `json:"last_err_code,omitempty"`
	LastErrMessage string `json:"last_err_message,omitempty"`
}
