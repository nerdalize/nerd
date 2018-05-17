package populator

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync/atomic"

	v1payload "github.com/nerdalize/nerd/nerd/client/auth/v1/payload"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd/api"
	// this blank import is necessary to load the oidc plugin for client-go
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

//GenericPopulator is an implementation of the P interface using on Open ID Connect credentials.
type GenericPopulator struct {
	// kubeConfigFile is the path where the kube config is stored
	// Only access this with atomic ops
	kubeConfigFile atomic.Value
	homedir        string
	cluster        *v1payload.GetClusterOutput
	client         *Client
}

func newGeneric(c *Client, kubeConfigFile, homedir string, cluster *v1payload.GetClusterOutput) *GenericPopulator {
	o := &GenericPopulator{
		cluster: cluster,
		homedir: homedir,
		client:  c,
	}
	o.kubeConfigFile.Store(kubeConfigFile)
	return o
}

//GetKubeConfigFile returns the path where the kube config is stored.
func (o *GenericPopulator) GetKubeConfigFile() string {
	return o.kubeConfigFile.Load().(string)
}

//RemoveConfig deletes the precised cluster context and cluster info.
func (o *GenericPopulator) RemoveConfig(cluster string) error {
	// read existing config or create new if does not exist
	kubecfg, err := ReadConfigOrNew(o.GetKubeConfigFile())
	if err != nil {
		return err
	}
	delete(kubecfg.Clusters, cluster)
	delete(kubecfg.AuthInfos, cluster)
	delete(kubecfg.Contexts, fmt.Sprintf("%s-%s", Prefix, cluster))
	kubecfg.CurrentContext = ""

	// write back to disk
	if err := WriteConfig(kubecfg, o.GetKubeConfigFile()); err != nil {
		return errors.Wrap(err, "could not write kubeconfig")
	}
	return nil
}

// PopulateKubeConfig populates an api.Config object and set the current context to the provided cluster.
func (o *GenericPopulator) PopulateKubeConfig(namespace string) error {
	c := api.NewCluster()
	if o.cluster == nil {
		return errors.New("Cannot use an empty cluster")
	}
	if o.cluster.CaCertificate == "" {
		c.InsecureSkipTLSVerify = true
	} else {
		cert, err := o.createCertificate(o.cluster.CaCertificate, o.cluster.ShortName, o.homedir)
		if err != nil {
			return err
		}
		c.CertificateAuthority = cert
	}
	c.Server = o.cluster.ServiceURL

	auth := api.NewAuthInfo()
	if o.cluster.KubeConfigUser.Token != "" {
		auth.Token = o.cluster.KubeConfigUser.Token
	} else {
		auth.AuthProvider = &api.AuthProviderConfig{
			Name: "oidc",
			Config: map[string]string{
				"client-id":      o.cluster.KubeConfigUser.AuthProvider.Config.ClientID,
				"id-token":       o.cluster.KubeConfigUser.AuthProvider.Config.IDToken,
				"idp-issuer-url": o.cluster.KubeConfigUser.AuthProvider.Config.IdpIssuerURL,
				"refresh-token":  o.cluster.KubeConfigUser.AuthProvider.Config.RefreshToken,
			},
		}
	}

	if namespace == "" {
		if len(o.cluster.Namespaces) != 0 {
			namespace = o.cluster.Namespaces[0].Name
		} else {
			namespace = "default"
		}
	}

	// context
	context := api.NewContext()
	context.Cluster = o.cluster.ShortName
	context.AuthInfo = o.cluster.ShortName
	context.Namespace = namespace
	clusterName := fmt.Sprintf("%s-%s", Prefix, o.cluster.ShortName)

	// read existing config or create new if does not exist
	kubecfg, err := ReadConfigOrNew(o.GetKubeConfigFile())
	if err != nil {
		return err
	}
	kubecfg.Clusters[o.cluster.ShortName] = c
	kubecfg.CurrentContext = clusterName
	kubecfg.AuthInfos[o.cluster.ShortName] = auth
	kubecfg.Contexts[clusterName] = context

	// write back to disk
	if err := WriteConfig(kubecfg, o.GetKubeConfigFile()); err != nil {
		return errors.Wrap(err, "could not write kubeconfig")
	}

	return nil
}

func (o *GenericPopulator) createCertificate(data, cluster, homedir string) (string, error) {
	if data == "" {
		return "", nil
	}
	dir := filepath.Join(homedir, ".nerd", "certs")
	filename := filepath.Join(dir, cluster+".cert")
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
