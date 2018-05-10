package v1payload

//ListClustersOutput contains a list of projects
type ListClustersOutput struct {
	Clusters []*GetClusterOutput
}

//GetClusterOutput get some details of a specific cluster. Useful to setup kube config.
type GetClusterOutput struct {
	URL           string `json:"url"`
	ShortName     string `json:"short_name"`
	Name          string `json:"name"`
	Namespace     string `json:"namespace"`
	ServiceType   string `json:"service_type"`
	ServiceURL    string `json:"service_url"`
	CaCertificate string `json:"ca_certificate"`
	Capacity      struct {
		CPU    string `json:"cpu"`
		Memory string `json:"memory"`
		Pods   int    `json:"pods"`
	} `json:"capacity"`
	Usage struct {
		CPU    string `json:"cpu"`
		Memory string `json:"memory"`
		Pods   int    `json:"pods"`
	} `json:"usage"`
}
