package payload

import "time"

//TaskCreateInput is used as input to task creation
type TaskCreateInput struct {
	Image       string            `json:"image" valid:"min=1,max=64,required"`
	InputID     string            `json:"input_id" valid:"datasetid"`
	Environment map[string]string `json:"environment"`
}

//TaskCreateOutput is returned from
type TaskCreateOutput struct {
	Task
}

//TaskListOutput is returned from the task listing
type TaskListOutput struct {
	Tasks []*Task `json:"tasks"`
}

//Task is a task in the list output
type Task struct {
	ProjectID   string            `json:"project_id"`
	CreatedAt   time.Time         `json:"created_at"`
	InputID     string            `json:"input_id"`
	OutputID    string            `json:"output_id"`
	Image       string            `json:"image"`
	Environment map[string]string `json:"environment"`
}
