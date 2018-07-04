package svc

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/nerdalize/nerd/pkg/kubevisor"
	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	crdbeta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	extscheme "k8s.io/apiextensions-apiserver/pkg/client/clientset/internalclientset/scheme"
	"k8s.io/client-go/kubernetes/scheme"
)

var (
	files = map[string]string{
		"custom-dataset-controller":      "https://raw.githubusercontent.com/nerdalize/catalog/master/templates/custom-dataset-controller.yaml",
		"kube-system-cluster-role":       "https://raw.githubusercontent.com/nerdalize/catalog/master/templates/kube-system-default.yaml",
		"kube-system-clusterrolebinding": "https://raw.githubusercontent.com/nerdalize/catalog/master/templates/kube-system-default-clusterrolebinding.yaml",
		"custom-dataset-definition":      "https://raw.githubusercontent.com/nerdalize/catalog/master/templates/custom-dataset-definition.yaml",
		"flexvolume-clusterrole":         "https://raw.githubusercontent.com/nerdalize/catalog/master/templates/flexvolume-clusterrole.yaml",
		"flexvolume-clusterrolebinding":  "https://raw.githubusercontent.com/nerdalize/catalog/master/templates/flexvolume-clusterrolebinding.yml",
		"flexvolume-daemonset":           "https://raw.githubusercontent.com/nerdalize/catalog/master/templates/flexvolume-daemonset.yaml",
	}
)

// AddNerdDependenciesInput is used to configure the resource creation
type AddNerdDependenciesInput struct {
	Dependencies []string

	// could be also used to specify a flexvolume version,
	// and to propagate s3 credentials so that people can use by default their own private s3 bucket.
}

// IsNerdCompliant checks if the nlz-utils are running on the current cluster
func (k *Kube) IsNerdCompliant(ctx context.Context) (ok bool, dependencies []string, err error) {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	for resource, url := range files {
		resp, err := netClient.Get(url)
		if err != nil {
			return false, nil, err
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return false, nil, err
		}

		var obj interface{}
		if resource == "custom-dataset-definition" {
			decode := extscheme.Codecs.UniversalDeserializer().Decode
			obj, _, err = decode(data, nil, nil)
		} else {
			decode := scheme.Codecs.UniversalDeserializer().Decode
			obj, _, err = decode(data, nil, nil)

		}
		if err != nil {
			return false, []string{}, fmt.Errorf("Error while decoding YAML object. Err was: %s", err)
		}

		switch o := obj.(type) {
		case *appsv1.Deployment:
			err = k.visor.GetResource(ctx, kubevisor.ResourceTypeDeployments, &appsv1.Deployment{}, o.Name)
		case *appsv1.DaemonSet:
			o.Namespace = "default"
			err = k.visor.GetResource(ctx, kubevisor.ResourceTypeDaemonsets, &appsv1.DaemonSet{}, o.Name)
		case *rbacv1.ClusterRole:
			err = k.visor.GetClusterResource(ctx, kubevisor.ResourceTypeClusterRoles, &rbacv1.ClusterRole{}, o.Name)
		case *rbacv1.ClusterRoleBinding:
			err = k.visor.GetClusterResource(ctx, kubevisor.ResourceTypeClusterRoleBindings, &rbacv1.ClusterRoleBinding{}, o.Name)
		case *crdbeta1.CustomResourceDefinition:
			err = k.visor.GetClusterResource(ctx, kubevisor.ResourceTypeCustomResourceDefinition, &crdbeta1.CustomResourceDefinition{}, o.Name)
		default:
			//o is unknown for us
		}
		if err != nil {
			if kubevisor.IsNotExistsErr(err) {
				dependencies = append(dependencies, resource)
			} else {
				return false, []string{}, err
			}
		}
	}
	if len(dependencies) == 0 {
		return true, dependencies, nil
	}
	return false, dependencies, nil
}

// AddNerdDependencies will deploy necessary daemonsets, controllers and roles so that a private cluster can be used by the cli
func (k *Kube) AddNerdDependencies(ctx context.Context, in *AddNerdDependenciesInput) (err error) {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}
	for _, dependency := range in.Dependencies {
		// Get the data
		resp, err := netClient.Get(files[dependency])
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

		var obj interface{}

		if dependency == "custom-dataset-definition" {
			decode := extscheme.Codecs.UniversalDeserializer().Decode
			obj, _, err = decode(data, nil, nil)
		} else {
			decode := scheme.Codecs.UniversalDeserializer().Decode
			obj, _, err = decode(data, nil, nil)

		}
		if err != nil {
			return fmt.Errorf("Error while decoding YAML object. Err was: %s", err)
		}
		// now use switch over the type of the object
		// and match each type-case

		switch o := obj.(type) {
		case *appsv1.Deployment:
			err = k.visor.CreateResource(ctx, kubevisor.ResourceTypeDeployments, o, o.Name)
		case *appsv1.DaemonSet:
			err = k.visor.CreateResource(ctx, kubevisor.ResourceTypeDaemonsets, o, o.Name)
		case *rbacv1.ClusterRole:
			err = k.visor.CreateClusterResource(ctx, kubevisor.ResourceTypeClusterRoles, o, o.Name)
		case *rbacv1.ClusterRoleBinding:
			err = k.visor.CreateClusterResource(ctx, kubevisor.ResourceTypeClusterRoleBindings, o, o.Name)
		case *crdbeta1.CustomResourceDefinition:
			err = k.visor.CreateClusterResource(ctx, kubevisor.ResourceTypeCustomResourceDefinition, o, o.Name)
		default:
			//o is unknown for us
		}
		if err != nil {
			return err
		}
	}
	return err
}
