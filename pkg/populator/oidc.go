package populator

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
	ClientSecret = "93177b0e77369537ceac900b26f0a9600484564fdda5d431b05e994b"
	// ClientID is a client id that all tokens must be issued for.
	ClientID = "T8I0H3qAeWGA"
	// IDPIssuerURL is the URL of the provider which allows the API server to discover public signing keys.
	IDPIssuerURL = "https://auth.nerdalize.com"
	// DirPermissions are the output directory's permissions.
	DirPermissions = 0755
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
	homedir string
}

func newOIDC(kubeConfigFile, homedir string, project *v1payload.GetProjectOutput) *OIDCPopulator {
	o := &OIDCPopulator{
		project: project,
		homedir: homedir,
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
		cert, err := o.createCertificate(o.project.Services.Cluster.B64CaData, project, o.homedir)
		if err != nil {
			return err
		}
		cluster.CertificateAuthority = cert
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
			"client-id":      ClientID,
			"client-secret":  ClientSecret,
			"id-token":       config.OAuth.IDToken,
			"idp-issuer-url": IDPIssuerURL,
			"refresh-token":  config.OAuth.RefreshToken,
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
	if data == "" {
		return "", nil
	}
	dir := filepath.Join(homedir, ".nerd", "certs")
	filename := filepath.Join(dir, project+".cert")
	_, err := os.Stat(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", errors.Errorf("'%v' is not a path", dir)
		}
		err = os.MkdirAll(dir, DirPermissions)
		if err != nil {
			return "", errors.Wrap(err, fmt.Sprintf("The provided path '%s' does not exist and could not be created.", dir))
		}
		_, err = os.Stat(dir)
		if err != nil {
			return "", err
		}
	}
	d, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(filename, d, 0644)
	if err != nil {
		return "", err
	}
	return filename, nil
}
