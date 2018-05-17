package v1payload

//ListClustersOutput contains a list of projects
type ListClustersOutput struct {
	Clusters []*GetClusterOutput
}

//GetClusterOutput get some details of a specific cluster. Useful to setup kube config.
type GetClusterOutput struct {
	URL        string `json:"url"`
	ShortName  string `json:"short_name"`
	Name       string `json:"name"`
	Namespaces []struct {
		Name string `json:"name"`
	} `json:"namespaces"`
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
	KubeConfigUser struct {
		BearerToken  string `json:"bearer_token"`
		AuthProvider struct {
			Config struct {
				IdpIssuerURL string `json:"idp-issuer-url"`
				ClientID     string `json:"client-id"`
				RefreshToken string `json:"refresh-token"`
				IDToken      string `json:"id-token"`
			} `json:"config"`
			Name string `json:"name"`
		} `json:"auth-provider"`
	} `json:"kubeconfig_user"`
}
