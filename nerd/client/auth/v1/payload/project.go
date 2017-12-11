package v1payload

//ListProjectsOutput contains a list of projects
type ListProjectsOutput struct {
	Projects []*Project
}

//GetProjectOutput get some details of a specific project. Useful to setup the right configuration.
type GetProjectOutput struct {
	ID         int    `json:"id"`
	ClusterURL string `json:"url"`
}

//Project represents a project
type Project struct {
	ID   int    `json:"id"`
	URL  string `json:"url"`
	Slug string `json:"slug"`
}
