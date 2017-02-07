package client

type REST struct {
}

type RESTConfig struct {
	Scheme   string
	Hostname string
	Port     string
	Version  string
}

func NewREST(config *RESTConfig) *REST {
	return &REST{}
}
