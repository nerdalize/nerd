package svc

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/nerdalize/nerd/pkg/kubevisor"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1beta1"
	crdbeta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	files = []string{
		"https://raw.githubusercontent.com/nerdalize/catalog/master/templates/kube-system-default.yaml",
		"https://raw.githubusercontent.com/nerdalize/catalog/master/templates/custom-dataset-controller.yaml",
		"https://raw.githubusercontent.com/nerdalize/catalog/master/templates/custom-dataset-definition.yaml",
		"https://raw.githubusercontent.com/nerdalize/catalog/master/templates/flexvolume-clusterrole.yaml",
		"https://raw.githubusercontent.com/nerdalize/catalog/master/templates/flexvolume-clusterrolebinding.yaml",
		"https://raw.githubusercontent.com/nerdalize/catalog/master/templates/flexvolume-daemonset.yaml",
	}
)

// NerdCompliantInput is used to configure the resource creation
type NerdCompliantInput struct {
	// could be used to specify a flexvolume version,
	// and to propagate s3 credentials so that people can use by default their own private s3 bucket.
}

// NerdCompliant will deploy necessary daemonsets, controllers and roles so that a private cluster can be used by the cli
func (k *Kube) NerdCompliant(ctx context.Context, in *NerdCompliantInput) (err error) {
	for _, url := range files {
		// Get the data
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		// data to yaml
		// pass config to kubernetes
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		obj, err := runtime.Decode(scheme.Codecs.UniversalDeserializer(), data)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error while decoding YAML object. Err was: %s", err))
		}

		// now use switch over the type of the object
		// and match each type-case
		switch o := obj.(type) {
		case *corev1.Pod:
			err = k.visor.CreateResource(ctx, kubevisor.ResourceTypePods, o, o.Name)
		case *appsv1.Deployment:
			err = k.visor.CreateResource(ctx, kubevisor.ResourceTypeDeployments, o, o.Name)
		case *appsv1.DaemonSet:
			err = k.visor.CreateResource(ctx, kubevisor.ResourceTypeDaemonsets, o, o.Name)
		case *v1beta1.Role:
			err = k.visor.CreateResource(ctx, kubevisor.ResourceTypeRoles, o, o.Name)
			// o is the actual role Object with all fields etc
		case *v1beta1.RoleBinding:
			err = k.visor.CreateResource(ctx, kubevisor.ResourceTypeRoleBindings, o, o.Name)
		case *v1beta1.ClusterRole:
			err = k.visor.CreateResource(ctx, kubevisor.ResourceTypeClusterRoles, o, o.Name)
		case *v1beta1.ClusterRoleBinding:
			err = k.visor.CreateResource(ctx, kubevisor.ResourceTypeClusterRoleBindings, o, o.Name)
		case *crdv1.Dataset:
			err = k.visor.CreateResource(ctx, kubevisor.ResourceTypeDatasets, o, o.Name)
		case *crdbeta1.CustomResourceDefinition:
			err = k.visor.CreateResource(ctx, kubevisor.ResourceTypeCustomResourceDefinition, o, o.Name)
		default:
			//o is unknown for us
		}
	}
	return err
}
