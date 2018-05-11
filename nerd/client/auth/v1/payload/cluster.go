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
	KubeConfigUser struct {
		BearerToken       string `json:"bearer_token"`
		AccessToken       string `json:"access_token"`
		RefreshToken      string `json:"refresh_token"`
		TokenType         string `json:"token_type"`
		ExpiresIn         int    `json:"expires_in"`
		IDToken           string `json:"id_token"`
		OauthClientID     string `json:"oauth_client_id"`
		OauthClientSecret string `json:"oauth_client_secret"`
	} `json:"kubeconfig_user"`
}
