package nerd

//Task describes a piece of workload
type Task struct {
	Image   string   `json:"image"`
	Dataset string   `json:"dataset"`
	Args    []string `json:"args"`
}

//TaskStatus is used to update the status of a task
type TaskStatus struct {
	Status string   `json:"status"`
	Logs   []string `json:"logs"`
}
