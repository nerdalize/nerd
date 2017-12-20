package v1payload

import "time"

//ListProjectsOutput contains a list of projects
type ListProjectsOutput struct {
	Projects []*Project
}

//GetProjectOutput get some details of a specific project. Useful to setup kube config.
type GetProjectOutput struct {
	URL      string `json:"url"`
	Nk       string `json:"nk"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Services struct {
		Cluster struct {
			URL        string `json:"url"`
			ID         int    `json:"id"`
			Name       string `json:"name"`
			Address    string `json:"address"`
			Subaddress string `json:"subaddress"`
			B64CaData  string `json:"b64_ca_data"`
			Default    bool   `json:"default"`
		} `json:"cluster"`
	} `json:"services"`
	CreatedAt time.Time `json:"created_at"`
}

//Project represents a project
type Project struct {
	ID   int    `json:"id"`
	URL  string `json:"url"`
	Slug string `json:"slug"`
}
