package payload

//Task is the structure used when we need to exchange with clients
type Task struct {
	ID       string   `json:"id"`
	Image    string   `json:"image"`
	Dataset  string   `json:"dataset"`
	Status   string   `json:"status"`
	Args     []string `json:"args"`
	LogLines []string `json:"log_lines"`
}

//TaskStatus is used to update the status of a task
type TaskStatus struct {
	Status string   `json:"status"`
	Logs   []string `json:"logs"`
}
