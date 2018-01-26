package populator

import (
	"fmt"
	"os"
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

//RemoveConfig deletes the precised project context and cluster info.
func (o *OIDCPopulator) RemoveConfig(project string) error {
	// read existing config or create new if does not exist
	kubecfg, err := ReadConfigOrNew(o.GetKubeConfigFile())
	if err != nil {
		return err
	}
	delete(kubecfg.Clusters, project)
	delete(kubecfg.AuthInfos, project)
	delete(kubecfg.Contexts, fmt.Sprintf("%s-%s", Prefix, project))
	kubecfg.CurrentContext = ""

	// write back to disk
	if err := WriteConfig(kubecfg, o.GetKubeConfigFile()); err != nil {
		return errors.Wrap(err, "could not write kubeconfig")
	}
	return nil
}

// PopulateKubeConfig populates an api.Config object and set the current context to the provided project.
func (o *OIDCPopulator) PopulateKubeConfig(project string) error {
	cluster := api.NewCluster()
	if o.project.Services.Cluster.B64CaData == "" {
		cluster.InsecureSkipTLSVerify = true
	} else {
		cluster.CertificateAuthorityData = []byte(o.project.Services.Cluster.B64CaData)
	}
	cluster.Server = o.project.Services.Cluster.Address

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
	clusterName := fmt.Sprintf("%s-%s", Prefix, project)

	// read existing config or create new if does not exist
	kubecfg, err := ReadConfigOrNew(o.GetKubeConfigFile())
	if err != nil {
		return err
	}
	kubecfg.Clusters[project] = cluster
	kubecfg.CurrentContext = clusterName
	kubecfg.AuthInfos[project] = auth
	kubecfg.Contexts[clusterName] = context

	// write back to disk
	if err := WriteConfig(kubecfg, o.GetKubeConfigFile()); err != nil {
		return errors.Wrap(err, "could not write kubeconfig")
	}

	return nil
}

func (o *OIDCPopulator) createCertificate(data, project, homedir string) (string, error) {
	fi, err := os.Stat(dataPath)
	if err != nil {
		return errors.Errorf("argument '%v' is not a valid file or directory", dataPath)
	}
	// check if certificate file exists
	// if not:
	// 	create file
	// 	decode b64 data
	//	write utf-8 data in file
	// close file
	// return path
	return "~/home/.kube/ca.pem", nil
}
