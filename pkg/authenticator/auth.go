//authenticator is a package that will help us to populate the kubernetes config file with the right credentials
package authenticator

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
)

// Configurer is an interface that we can use to read from and to write to the kube config file.
type Configurer interface {
	AuthGetter(userID int64, projects []string) (map[string]*api.Cluster, map[string]*api.Context, map[string]*api.AuthInfo)
	GetNamespace() string
	PopulateKubeConfig(serverName, serverAddress, namespace, user *api.AuthInfo) error
	SetKubeConfigFile(kubeConfigFile string)
	GetKubeConfigFile() string
}

// PopulateKubeConfigFromEnv populates the kube config file with the info found in the environment.
func PopulateKubeConfigFromEnv(project, kubeConfigFile string) error {
	cluster := api.NewCluster()
	cluster.Server = os.Getenv("KUBE_CLUSTER_ADDR")

	// user
	user := api.NewAuthInfo()
	user.Token = os.Getenv("KUBE_TOKEN")

	// context
	context := api.NewContext()
	context.Cluster = project
	context.AuthInfo = project
	context.Namespace = os.Getenv("KUBE_NAMESPACE")

	// read existing config or create new if does not exist
	kubecfg, err := ReadConfigOrNew(kubeConfigFile)
	if err != nil {
		return err
	}
	kubecfg.Clusters[project] = cluster
	kubecfg.AuthInfos[project] = user
	kubecfg.CurrentContext = project
	kubecfg.Contexts[project] = context

	// write back to disk
	if err := WriteConfig(kubecfg, kubeConfigFile); err != nil {
		return errors.Wrap(err, "writing kubeconfig")
	}

	return nil
}

// PopulateKubeConfig populates an api.Config object.
func PopulateKubeConfig(project, token, refresh_token, kubeConfigFile string) error {
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
	kubecfg, err := ReadConfigOrNew(kubeConfigFile)
	if err != nil {
		return err
	}
	kubecfg.Clusters[project] = cluster
	kubecfg.AuthInfos[project] = user
	kubecfg.CurrentContext = project
	kubecfg.Contexts[project] = context

	// write back to disk
	if err := WriteConfig(kubecfg, kubeConfigFile); err != nil {
		return errors.Wrap(err, "writing kubeconfig")
	}

	return nil
}

// ReadConfigOrNew retrieves Kubernetes client configuration from a file.
// If no files exists, an empty configuration is returned.
func ReadConfigOrNew(filename string) (*api.Config, error) {
	data, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		return api.NewConfig(), nil
	} else if err != nil {
		return nil, errors.Wrapf(err, "Error reading file %q", filename)
	}

	// decode config, empty if no bytes
	config, err := decode(data)
	if err != nil {
		return nil, errors.Errorf("could not read config: %v", err)
	}

	// initialize nil maps
	if config.AuthInfos == nil {
		config.AuthInfos = map[string]*api.AuthInfo{}
	}
	if config.Clusters == nil {
		config.Clusters = map[string]*api.Cluster{}
	}
	if config.Contexts == nil {
		config.Contexts = map[string]*api.Context{}
	}

	return config, nil
}

// WriteConfig encodes the configuration and writes it to the given file.
// If the file exists, it's contents will be overwritten.
func WriteConfig(config *api.Config, filename string) error {
	if config == nil {
		glog.Errorf("could not write to '%s': config can't be nil", filename)
	}

	// encode config to YAML
	data, err := runtime.Encode(latest.Codec, config)
	if err != nil {
		return errors.Errorf("could not write to '%s': failed to encode config: %v", filename, err)
	}

	// create parent dir if doesn't exist
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrapf(err, "Error creating directory: %s", dir)
		}
	}

	// write with restricted permissions
	if err := ioutil.WriteFile(filename, data, 0600); err != nil {
		return errors.Wrapf(err, "Error writing file %s", filename)
	}
	// if err := util.MaybeChownDirRecursiveToMinikubeUser(dir); err != nil {
	// 	return errors.Wrapf(err, "Error recursively changing ownership for dir: %s", dir)
	// }

	return nil
}

//Namespace get the namespace of the current context from the kube config file
func Namespace(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if os.IsNotExist(err) {
		return "", err
	} else if err != nil {
		return "", errors.Wrapf(err, "Error reading file %q", filename)
	}

	// decode config, empty if no bytes
	config, err := decode(data)
	if err != nil {
		return "", errors.Errorf("could not read config: %v", err)
	}

	fmt.Println(config.Contexts[config.CurrentContext].Namespace, config.CurrentContext)
	return config.Contexts[config.CurrentContext].Namespace, nil
}

// decode reads a Config object from bytes.
// Returns empty config if no bytes.
func decode(data []byte) (*api.Config, error) {
	// if no data, return empty config
	if len(data) == 0 {
		return api.NewConfig(), nil
	}

	config, _, err := latest.Codec.Decode(data, nil, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "Error decoding config from data: %s", string(data))
	}

	return config.(*api.Config), nil
}
