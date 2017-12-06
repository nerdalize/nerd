package authenticator

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/clientcmd/api/latest"
)

type ConfigFile interface {
	Get() *Config
	Populate(conf *Config) error
}

type Config interface {
	GetAuth(userID int64, projects []string) (map[string]*api.Cluster, map[string]*api.Context, map[string]*api.AuthInfo)
	GetClusters() map[string]*api.Cluster
	GetContext(context string) *api.Context
	GetUsers() *api.AuthInfo
	SetCluster(serverName, serverAddress string)
	SetContext(clusterName, namespace, username string)
	SetUser(info *api.AuthInfo)
}

type FromEnv struct {
}

type KubeConfigSetup struct {
	// The name of the cluster for this context
	ClusterName string

	// ClusterServerAddress is the address of of the kubernetes cluster
	ClusterServerAddress string

	Username  string
	Token     string
	Namespace string

	// // ClientCertificate is the path to a client cert file for TLS.
	// ClientCertificate string

	// // CertificateAuthority is the path to a cert file for the certificate authority.
	// CertificateAuthority string

	// // ClientKey is the path to a client key file for TLS.
	// ClientKey string

	// Should the current context be kept when setting up this one
	// KeepContext bool

	// kubeConfigFile is the path where the kube config is stored
	// Only access this with atomic ops
	kubeConfigFile atomic.Value
}

func (k *KubeConfigSetup) SetKubeConfigFile(kubeConfigFile string) {
	k.kubeConfigFile.Store(kubeConfigFile)
}

func (k *KubeConfigSetup) GetKubeConfigFile() string {
	return k.kubeConfigFile.Load().(string)
}

// PopulateKubeConfig populates an api.Config object.
func PopulateKubeConfig(cfg *KubeConfigSetup, kubecfg *api.Config) {
	clusterName := cfg.ClusterName
	cluster := api.NewCluster()
	cluster.Server = cfg.ClusterServerAddress
	kubecfg.Clusters[clusterName] = cluster

	// user
	userName := cfg.ClusterName
	user := api.NewAuthInfo()
	user.Username = cfg.Username
	user.Token = cfg.Token
	kubecfg.AuthInfos[userName] = user

	// context
	contextName := cfg.ClusterName
	context := api.NewContext()
	context.Cluster = cfg.ClusterName
	context.AuthInfo = userName
	context.Namespace = cfg.Namespace
	kubecfg.CurrentContext = contextName
	kubecfg.Contexts[contextName] = context
}

// SetupKubeConfig reads config from disk, adds the minikube settings, and writes it back.
// activeContext is true when minikube is the CurrentContext
// If no CurrentContext is set, the given name will be used.
func SetupKubeConfig(cfg *KubeConfigSetup) error {
	glog.Infof("Using kubeconfig: ", cfg.GetKubeConfigFile())

	// read existing config or create new if does not exist
	config, err := ReadConfigOrNew(cfg.GetKubeConfigFile())
	if err != nil {
		return err
	}

	PopulateKubeConfig(cfg, config)
	fmt.Println(config.Contexts[config.CurrentContext].Namespace)

	// write back to disk
	if err := WriteConfig(config, cfg.GetKubeConfigFile()); err != nil {
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

func (f *FromEnv) AuthGetter(userID int64, projects []string) (map[string]*api.Cluster, map[string]*api.Context, map[string]*api.AuthInfo) {
	// get kubeCfgSetup from authentication service (token?)
	kubeCfgSetup := &KubeConfigSetup{
		ClusterName:          os.Getenv("CLUSTER_NAME"),
		ClusterServerAddress: os.Getenv("CLUSTER_ADDRESS"), // get it from project
		Username:             os.Getenv("CLUSTER_NAME"),
		Token:                os.Getenv("CLUSTER_TOKEN"),
		Namespace:            project,
	}
	kubeCfgSetup.SetKubeConfigFile(kubeConfigFile)

	if err := SetupKubeConfig(kubeCfgSetup); err != nil {
		glog.Errorln("Error setting up kubeconfig: ", err)
	}
}
