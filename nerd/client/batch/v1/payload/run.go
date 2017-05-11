package v1payload

//An Run acts as an reference to a task instance
type Run struct {
	ProjectID string `json:"project_id"`
	QueueID   string `json:"queue_id"`
	TaskID    int64  `json:"task_id"`
	Token     string `json:"token"`

	Cmd   []string          `json:"cmd"`
	Env   map[string]string `json:"env"`
	Stdin []byte            `json:"stdin"`
}

//SendRunHeartbeatInput is input for queue creation
type SendRunHeartbeatInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	QueueID   string `json:"queue_id" valid:"required"`
	TaskID    int64  `json:"task_id" valid:"required"`
	RunToken  string `json:"run_token" valid:"required"`
}

//SendRunHeartbeatOutput is output for queue creation
type SendRunHeartbeatOutput struct {
	HasExpired bool `json:"has_expired"`
}

//SendRunSuccessInput is input for marking a run as failed
type SendRunSuccessInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	QueueID   string `json:"queue_id" valid:"required"`
	TaskID    int64  `json:"task_id" valid:"required"`
	RunToken  string `json:"run_token" valid:"required"`

	Result string `json:"result"`
}

//SendRunSuccessOutput is output from marking a run as failed
type SendRunSuccessOutput struct{}

//SendRunFailureInput is input for marking a run as failed
type SendRunFailureInput struct {
	ProjectID string `json:"project_id" valid:"required"`
	QueueID   string `json:"queue_id" valid:"required"`
	TaskID    int64  `json:"task_id" valid:"required"`
	RunToken  string `json:"run_token" valid:"required"`

	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
}

//SendRunFailureOutput is output from marking a run as failed
type SendRunFailureOutput struct{}
