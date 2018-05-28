package kubevisor

import (
	crd "github.com/nerdalize/nerd/crd/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

//ResourceType is a type of Kubernetes resource
type ResourceType string

var (
	//ResourceTypeJobs is used for job management
	ResourceTypeJobs = ResourceType("jobs")

	//ResourceTypePods is used for pod inspection
	ResourceTypePods = ResourceType("pods")

	//ResourceTypeDatasets is used for dataset management
	ResourceTypeDatasets = ResourceType("datasets")

	//ResourceTypeEvents is the resource type for event fetching
	ResourceTypeEvents = ResourceType("events")

	//ResourceTypeQuota can be used to retrieve quota information
	ResourceTypeQuota = ResourceType("resourcequotas")

	//ResourceTypeSecrets can be used to get secret information
	ResourceTypeSecrets = ResourceType("secrets")

	//ResourceTypeDeployment is used for deployment management
	ResourceTypeDeployments = ResourceType("deployments")

	//ResourceTypeRoles is used for role management
	ResourceTypeRoles = ResourceType("roles")

	//ResourceTypeRoleBindings is used for role bindings management
	ResourceTypeRoleBindings = ResourceType("rolebindings")

	//ResourceTypeClusterRoles is used for cluster roles management
	ResourceTypeClusterRoles = ResourceType("clusterroles")

	//ResourceTypeClusterRoleBindings is used for cluster role bindings management
	ResourceTypeClusterRoleBindings = ResourceType("clusterrolebindings")

	//ResourceTypeDaemonsets is used for daemonset management
	ResourceTypeDaemonsets = ResourceType("daemonsets")

	//ResourceTypeCustomResourceDefinition is used for crd management
	ResourceTypeCustomResourceDefinitions = ResourceType("customresourcedefintions")
)

//ManagedNames allows for Nerd to transparently manage resources based on names and there prefixes
type ManagedNames interface {
	GetName() string
	GetLabels() map[string]string
	SetLabels(map[string]string)
	SetName(name string)
	SetGenerateName(name string)
}

//ListTranformer must be implemented to allow Nerd to transparently manage resource names
type ListTranformer interface {
	Transform(fn func(in ManagedNames) (out ManagedNames))
	Len() int
}

//Visor provides access to Kubernetes resources while transparently filtering, naming and labeling
//resources that are managed by the CLI.
type Visor struct {
	prefix string
	ns     string
	api    kubernetes.Interface
	crd    crd.Interface
	logs   Logger
}
