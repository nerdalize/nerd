//authenticator is a package that will help us to populate the kubernetes config file with the right credentials
package authenticator

import (
	"k8s.io/client-go/tools/clientcmd/api"
)

type ConfigFile interface {
	Get() *Config
	Populate(conf *Config) error
}

type Config interface {
	AuthGetter(userID int64, projects []string) (map[string]*api.Cluster, map[string]*api.Context, map[string]*api.AuthInfo)
	GetClusters() map[string]*api.Cluster
	GetContext(context string) *api.Context
	GetUsers() *api.AuthInfo
	SetCluster(serverName, serverAddress string)
	SetContext(clusterName, namespace, username string)
	SetUser(info *api.AuthInfo)
}
