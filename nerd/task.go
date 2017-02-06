package nerd

//Task describes a piece of workload
type Task struct {
	ID       string   `json:"id"`
	Image    string   `json:"image"`
	Dataset  string   `json:"dataset"`
	Args     []string `json:"args"`
	Status   string   `json:"status"`
	LogLines []string `json:"log_lines"`
}

//TaskStatus is used to update the status of a task
type TaskStatus struct {
	Status string   `json:"status"`
	Logs   []string `json:"logs"`
}
