package populator

import (
	"fmt"
	"sync/atomic"

	"github.com/nerdalize/nerd/nerd/conf"

	v1payload "github.com/nerdalize/nerd/nerd/client/auth/v1/payload"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd/api"
	// this blank import is necessary to load the oidc plugin for client-go
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

const (
	// ClientSecret necessary for OpenID connect
	ClientSecret = "f9ef9cb57f5a76e0715def8e7c4c609a1b8872912bc09208cb75d71f"
	// ClientID is a client id that all tokens must be issued for.
	ClientID = "ckvyq40yyGSH"
	// IDPIssuerURL is the URL of the provider which allows the API server to discover public signing keys.
	IDPIssuerURL = "https://auth.nerdalize.com"
)

var (
	authEndpoint  = fmt.Sprintf("%s/v1/o/authorize", IDPIssuerURL)
	tokenEndpoint = fmt.Sprintf("%s/v1/o/token", IDPIssuerURL)
)

//OIDCPopulator is an implementation of the P interface using on Open ID Connect credentials.
type OIDCPopulator struct {
	// kubeConfigFile is the path where the kube config is stored
	// Only access this with atomic ops
	kubeConfigFile atomic.Value

	project *v1payload.GetProjectOutput
}

func newOIDC(kubeConfigFile string, project *v1payload.GetProjectOutput) *OIDCPopulator {
	o := &OIDCPopulator{
		project: project,
	}
	o.kubeConfigFile.Store(kubeConfigFile)
	return o
}

//GetKubeConfigFile returns the path where the kube config is stored.
func (o *OIDCPopulator) GetKubeConfigFile() string {
	return o.kubeConfigFile.Load().(string)
}

// PopulateKubeConfig populates an api.Config object and set the current context to the provided project.
func (o *OIDCPopulator) PopulateKubeConfig(project string) error {
	cluster := api.NewCluster()
	cluster.InsecureSkipTLSVerify = true
	cluster.Server = o.project.Services.Cluster.Address
	if cluster.Server == "" {
		return errors.New("this project isn't available for open id connect (server address cannot be blank)")
	}

	filename, err := conf.GetDefaultSessionLocation()
	if err != nil {
		return err
	}
	ss := conf.NewSession(filename)
	if err != nil {
		return err
	}
	config, err := ss.Read()
	if err != nil {
		return err
	}

	auth := api.NewAuthInfo()
	auth.AuthProvider = &api.AuthProviderConfig{
		Name: "oidc",
		Config: map[string]string{
			"client-id":                 ClientID,
			"client-secret":             ClientSecret,
			"id-token":                  config.OAuth.IDToken,
			"idp-certificate-authority": o.project.Services.Cluster.B64CaData,
			"idp-issuer-url":            IDPIssuerURL,
			"refresh-token":             config.OAuth.RefreshToken,
		},
	}

	// context
	context := api.NewContext()
	context.Cluster = project
	context.AuthInfo = project
	context.Namespace = project

	// read existing config or create new if does not exist
	kubecfg, err := ReadConfigOrNew(o.GetKubeConfigFile())
	if err != nil {
		return err
	}
	kubecfg.Clusters[project] = cluster
	kubecfg.CurrentContext = project
	kubecfg.AuthInfos[project] = auth
	kubecfg.Contexts[project] = context

	// write back to disk
	if err := WriteConfig(kubecfg, o.GetKubeConfigFile()); err != nil {
		return errors.Wrap(err, "writing kubeconfig")
	}

	return nil
}
