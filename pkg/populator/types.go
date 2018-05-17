package populator

const (
	// DirPermissions are the output directory's permissions.
	DirPermissions = 0755

	//Prefix is used to know if a context comes from the cli.
	Prefix = "nerd-cli"
)

// P is an interface that we can use to read from and to write to the kube config file.
type P interface {
	PopulateKubeConfig(string) error
	RemoveConfig(string) error
}

// Client provides necessary information to successfully use OIDC
type Client struct {
	// TODO CLEANING
	// Secret necessary for OpenID connect
	Secret string
	// ID is a client id that all tokens must be issued for.
	ID string
	// IDPIssuerURL is the URL of the provider which allows the API server to discover public signing keys.
	IDPIssuerURL string
}

// Context information
type Context struct {
	Name      string `long:"context" description:"context to use for this configuration"`
	Namespace string `long:"context-ns" description:"" default:"default"`
	Server    string `long:"context-server" description:"if not provided, will use the cluster name"`
	User      string `long:"context-user" description:" If not provided, will use the info from the next options"`
}

// Cluster information
type Cluster struct {
	Name   string `long:"cluster" description:"name of the cluster configuration entry"`
	Server string `long:"server" short:"s" description:"url of the cluster to reach"`
	CA     string `long:"certificate-authority" description:""`
	CAPath string `long:"certificate-authority-path" description:""`
}

// Auth information
type Auth struct {
	User     string `long:"user" description:"user to use for this configuration"`
	Password string `long:"password" description:"password for the user entry in kubeconfig"`
	Token    string `long:"token" description:"token for the user entry in kubeconfig"`
	Username string `long:"username" description:"username for the user entry in kubeconfig"`

	AuthProvider      string   `long:"auth-provider" description:""`
	AuthProviderArgs  []string `long:"auth-provider-arg" description:"key=value"`
	ClientCertificate string   `long:"client-certificate" description:"Path to client-certificate file for the user entry in kubeconfig"`
	ClientKey         string   `long:"client-key" description:"Path to client-key file for the user entry in kubeconfig"`
	EmbedCerts        string   `long:"embed-certs" description:"Embed client cert/key for the user entry in kubeconfig"`

	// OIDC
	// kubectl config set-credentials USER_NAME \
	// 	--auth-provider=oidc \
	// 	--auth-provider-arg=idp-issuer-url=( issuer url ) \
	// 	--auth-provider-arg=client-id=( your client id ) \
	// 	--auth-provider-arg=client-secret=( your client secret ) \
	// 	--auth-provider-arg=refresh-token=( your refresh token ) \
	// 	--auth-provider-arg=idp-certificate-authority=( path to your ca certificate ) \
	// 	--auth-provider-arg=id-token=( your id_token )

	SecureClientSecret string `long:"secure-client-secret" description:"mandatory to setup OIDC"`
	SecureClientID     string `long:"secure-client-id" description:"mandatory to setup OIDC"`
	IDPIssuerURL       string `long:"idp-issuer-url" description:"mandatory to setup OIDC"`
	IDPCaPath          string `long:"idp-ca-path" description:"path to your ca certificate"`
	RefreshToken       string `long:"refresh-token" description:"mandatory to setup OIDC"`
	IDToken            string `long:"id-token" description:"mandatory to setup OIDC"`
}
