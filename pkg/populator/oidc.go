package populator

import (
	"os"
	"sync/atomic"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	// ClientSecret necessary for OpenID connect
	ClientSecret = "0c4feb1e9d11790451a4364e803284a60905cef1a5f9bf7bad5f0eeb"
	// ClientID is a client id that all tokens must be issued for.
	ClientID = "myclientid"
	// IDPIssuerURL is the URL of the provider which allows the API server to discover public signing keys.
	IDPIssuerURL = "https://oidc.nce.nerdalize.com/v1/o"
)

type OIDCPopulator struct {
	// kubeConfigFile is the path where the kube config is stored
	// Only access this with atomic ops
	kubeConfigFile atomic.Value
}

func newOIDC(kubeConfigFile string) *OIDCPopulator {
	o := &OIDCPopulator{}
	o.kubeConfigFile.Store(kubeConfigFile)
	return o
}

func (o *OIDCPopulator) GetKubeConfigFile() string {
	return o.kubeConfigFile.Load().(string)
}

// PopulateKubeConfig populates an api.Config object and set the current context to the provided project.
func (o *OIDCPopulator) PopulateKubeConfig(project string) error {
	cluster := api.NewCluster()
	cluster.Server = os.Getenv("KUBE_CLUSTER_ADDR")

	// user
	user := api.NewAuthInfo()
	user.Username = project
	user.Token = os.Getenv("KUBE_TOKEN")

	// context
	context := api.NewContext()
	context.Cluster = project
	context.AuthInfo = project
	context.Namespace = os.Getenv("KUBE_NAMESPACE")

	// read existing config or create new if does not exist
	kubecfg, err := ReadConfigOrNew(o.GetKubeConfigFile())
	if err != nil {
		return err
	}
	kubecfg.Clusters[project] = cluster
	kubecfg.AuthInfos[project] = user
	kubecfg.CurrentContext = project
	kubecfg.Contexts[project] = context

	// write back to disk
	if err := WriteConfig(kubecfg, o.GetKubeConfigFile()); err != nil {
		return errors.Wrap(err, "writing kubeconfig")
	}

	return nil
}
