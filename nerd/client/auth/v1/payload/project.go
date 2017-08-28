package v1payload

//ListProjectsOutput contains a list of projects
type ListProjectsOutput struct {
	Projects []*Project
}

//Project represents a project
type Project struct {
	ID   int    `json:"id"`
	URL  string `json:"url"`
	Slug string `json:"slug"`
}
