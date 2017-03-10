package payload

import "time"

//TaskCreateInput is used as input to task creation
type TaskCreateInput struct {
	Image       string            `json:"image" valid:"min=1,max=64,required"`
	InputID     string            `json:"input_id" valid:"datasetid"`
	Environment map[string]string `json:"environment"`
}

//TaskResult is used when the worker needs to provide results of the execution
type TaskResult struct {
	ProjectID  string `json:"project_id"`
	TaskID     string `json:"task_id"`
	OutputID   string `json:"output_id"`
	ExitStatus string `json:"exit_status"`
}

//TaskCreateOutput is returned from
type TaskCreateOutput struct {
	Task
}

//TaskDescribeOutput is returned from a specific task
type TaskDescribeOutput struct {
	Task
}

//TaskListOutput is returned from the task listing
type TaskListOutput struct {
	Tasks []*TaskSummary `json:"tasks"`
}

//TaskSummary is a summarized view of a task
type TaskSummary struct {
	ProjectID string    `json:"project_id"`
	TaskID    string    `json:"task_id"`
	OutputID  string    `json:"output_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

//Task is a task in the list output
type Task struct {
	ProjectID     string            `json:"project_id"`
	TaskID        string            `json:"task_id"`
	InputID       string            `json:"input_id"`
	OutputID      string            `json:"output_id,omitempty"`
	WorkerID      string            `json:"worker_id,omitempty"`
	Image         string            `json:"image"`
	Environment   map[string]string `json:"environment,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	ActivityToken string            `json:"activity_token,omitempty"`
}
