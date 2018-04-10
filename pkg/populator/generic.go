package populator

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/nerdalize/nerd/nerd/conf"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd/api"
)

//GenericPopulator is an implementation of the P interface using on Open ID Connect credentials.
type GenericPopulator struct {
	// kubeConfigFile is the path where the kube config is stored
	// Only access this with atomic ops
	kubeConfigFile atomic.Value

	homedir string
}

func newGeneric(kubeConfigFile, homedir string) *GenericPopulator {
	o := &GenericPopulator{
		homedir: homedir,
	}
	o.kubeConfigFile.Store(kubeConfigFile)
	return o
}

//GetKubeConfigFile returns the path where the kube config is stored.
func (o *GenericPopulator) GetKubeConfigFile() string {
	return o.kubeConfigFile.Load().(string)
}

//RemoveConfig deletes the precised project context and cluster info.
func (o *GenericPopulator) RemoveConfig(project string) error {
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
func (o *GenericPopulator) PopulateKubeConfig(project string) error {
	cluster := api.NewCluster()
	cluster.Server = ""

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

	_ = config

	auth := api.NewAuthInfo()
	auth.AuthProvider = &api.AuthProviderConfig{}

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

func (o *GenericPopulator) createCertificate(data, project, homedir string) (string, error) {
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
