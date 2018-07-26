package kubevisor

import (
	"context"
	"io"
	"net"
	"net/url"
	"strings"

	crd "github.com/nerdalize/nerd/crd/pkg/client/clientset/versioned"
	crdscheme "github.com/nerdalize/nerd/crd/pkg/client/clientset/versioned/scheme"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

var (
	//MaxLogBytes determines how much logs we're gonna return, we cap it hard at this point
	MaxLogBytes = int64(1024 * 1024) //1MiB
	//DefaultPrefix is used to identify a job created by the cli
	DefaultPrefix = "nlz-nerd"
)

//Logger describes the logging dependency the services require
type Logger interface {
	Debugf(format string, args ...interface{})
}

var (
	//used during deletion but requires an address to create a pointer for
	deletePropagationForeground = metav1.DeletePropagationForeground
)

//NewVisor will setup a Kubernetes visor
func NewVisor(ns, prefix string, api kubernetes.Interface, crd crd.Interface, apiext apiext.Interface, logs Logger) *Visor {
	if prefix == "" {
		prefix = DefaultPrefix
	}
	return &Visor{prefix, ns, api, crd, apiext, logs}
}

func (k *Visor) hasPrefix(n string) bool {
	return strings.HasPrefix(n, k.prefix)
}

func (k *Visor) applyPrefix(n string) string {
	return k.prefix + n
}

func (k *Visor) removePrefix(n string) string {
	return strings.TrimPrefix(n, k.prefix)
}

//GetResource will use the kube RESTClient to describe a resource by its name.
func (k *Visor) GetResource(ctx context.Context, t ResourceType, v ManagedNames, name string) (err error) {
	vv, ok := v.(runtime.Object)
	if !ok {
		return errors.Errorf("provided value was not castable to runtime.Object")
	}

	var c rest.Interface
	switch t {
	case ResourceTypeJobs:
		c = k.api.BatchV1().RESTClient()
	case ResourceTypeSecrets, ResourceTypePersistentVolumeClaims:
		c = k.api.CoreV1().RESTClient()
	case ResourceTypeDatasets:
		c = k.crd.NerdalizeV1().RESTClient()
	case ResourceTypeDaemonsets, ResourceTypeDeployments:
		c = k.api.AppsV1().RESTClient()
	default:
		return errors.Errorf("unknown Kubernetes resource type provided: '%s'", t)
	}

	name = k.applyPrefix(name)

	k.logs.Debugf("getting %s '%s' in namespace '%s': %s", t, name, k.ns, ctx)
	err = c.Get().
		Name(name).
		Namespace(k.ns).
		Resource(string(t)).
		Body(vv).
		Context(ctx).
		Do().
		Into(vv)

	if err != nil {
		return k.tagError(err)
	}

	v.SetName(k.removePrefix(v.GetName())) //normalize back to unprefixed resource name
	return nil
}

//GetClusterResource will use the kube RESTClient to describe a resource by its name.
func (k *Visor) GetClusterResource(ctx context.Context, t ResourceType, v ManagedNames, name string) (err error) {
	vv, ok := v.(runtime.Object)
	if !ok {
		return errors.Errorf("provided value was not castable to runtime.Object")
	}

	var c rest.Interface
	switch t {
	case ResourceTypeCustomResourceDefinition:
		c = k.apiext.ApiextensionsV1beta1().RESTClient()
	case ResourceTypeRoles, ResourceTypeRoleBindings, ResourceTypeClusterRoles, ResourceTypeClusterRoleBindings:
		c = k.api.RbacV1().RESTClient()
	default:
		return errors.Errorf("unknown Kubernetes resource type provided: '%s'", t)
	}

	k.logs.Debugf("getting %s '%s' in namespace '%s': %s", t, name, k.ns, ctx)
	err = c.Get().
		Name(name).
		Resource(string(t)).
		Body(vv).
		Context(ctx).
		Do().
		Into(vv)

	if err != nil {
		return k.tagError(err)
	}

	v.SetName(k.removePrefix(v.GetName())) //normalize back to unprefixed resource name
	return nil
}

//DeleteResource will use the kube RESTClient to delete a resource by its name.
func (k *Visor) DeleteResource(ctx context.Context, t ResourceType, name string) (err error) {
	var c rest.Interface
	switch t {
	case ResourceTypeJobs:
		c = k.api.BatchV1().RESTClient()
	case ResourceTypeDatasets:
		c = k.crd.NerdalizeV1().RESTClient()
	case ResourceTypeSecrets, ResourceTypePersistentVolumeClaims:
		c = k.api.CoreV1().RESTClient()
	default:
		return errors.Errorf("unknown Kubernetes resource type provided for deletion: '%s'", t)
	}

	name = k.applyPrefix(name)

	k.logs.Debugf("deleting %s '%s' in namespace '%s': %s", t, name, k.ns, ctx)
	err = c.Delete().
		Namespace(k.ns).
		Body(&metav1.DeleteOptions{PropagationPolicy: &deletePropagationForeground}).
		Resource(string(t)).
		Name(name).
		Context(ctx).
		Do().Error()

	if err != nil {
		return k.tagError(err)
	}

	return nil
}

//FetchLogs will read logs from container with name 'cname' from pod 'pname' and write it to writer 'w'
func (k *Visor) FetchLogs(ctx context.Context, tail int64, w io.Writer, cname, pname string) (err error) {
	taill := &tail
	if tail < 1 {
		taill = nil
	}

	pname = k.applyPrefix(pname)
	req := k.api.CoreV1().Pods(k.ns).GetLogs(pname, &corev1.PodLogOptions{
		Container:  cname,
		TailLines:  taill, //can be nil if we dont want to tail
		LimitBytes: &MaxLogBytes,
	})

	req = req.Context(ctx)
	rc, err := req.Stream()
	if err != nil {
		return k.tagError(err)
	}

	defer rc.Close()
	_, err = io.Copy(w, rc)
	if err != nil {
		return errors.Wrap(err, "failed to copy logs over")
	}

	return nil
}

//CreateResource will use the kube RESTClient to create a resource while using the context, adding the
//Nerd prefix and handling errors specific to our domain.
func (k *Visor) CreateResource(ctx context.Context, t ResourceType, v ManagedNames, name string) (err error) {
	vv, ok := v.(runtime.Object)
	if !ok {
		return errors.Errorf("provided value was not castable to runtime.Object")
	}

	genfix := "x-"
	var c rest.Interface
	switch t {
	case ResourceTypeJobs:
		c = k.api.BatchV1().RESTClient()
		genfix = "j-"
	case ResourceTypePods:
		c = k.api.CoreV1().RESTClient()
		genfix = "p-"
	case ResourceTypeDeployments, ResourceTypeDaemonsets:
		c = k.api.AppsV1().RESTClient()
		genfix = "d-"
	case ResourceTypeSecrets:
		c = k.api.CoreV1().RESTClient()
		genfix = "s-"
	case ResourceTypeDatasets:
		c = k.crd.NerdalizeV1().RESTClient()
		genfix = "d-"
        case ResourceTypePersistentVolumeClaims:
		c = k.api.CoreV1().RESTClient()
		genfix = "f-"
	default:
		return errors.Errorf("unknown Kubernetes resource type provided for creation: '%s'", t)
	}

	if name != "" && genfix != "" {
		v.SetName(k.applyPrefix(name))
	} else {
		v.SetGenerateName(k.prefix + genfix)
	}

	labels := v.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	labels["nerd-app"] = "cli"
	v.SetLabels(labels)

	k.logs.Debugf("creating %s '%s' in namespace '%s' and labels '%v': %s", t, v.GetName(), k.ns, labels, ctx)
	err = c.Post().
		Namespace(k.ns).
		Resource(string(t)).
		Body(vv).
		Context(ctx).
		Do().
		Into(vv)

	if err != nil {
		return k.tagError(err)
	}

	v.SetName(k.removePrefix(v.GetName())) //normalize back to unprefixed resource name
	return nil
}

//CreateClusterResource will use the kube RESTClient to create a cluster resource while using the context, adding the
//Nerd prefix and handling errors specific to our domain.
func (k *Visor) CreateClusterResource(ctx context.Context, t ResourceType, v ManagedNames, name string) (err error) {
	vv, ok := v.(runtime.Object)
	if !ok {
		return errors.Errorf("provided value was not castable to runtime.Object")
	}

	var c rest.Interface
	switch t {
	case ResourceTypeCustomResourceDefinition:
		c = k.apiext.ApiextensionsV1beta1().RESTClient()
	case ResourceTypeRoles, ResourceTypeRoleBindings, ResourceTypeClusterRoles, ResourceTypeClusterRoleBindings:
		c = k.api.RbacV1().RESTClient()
	default:
		return errors.Errorf("unknown Kubernetes resource type provided for creation: '%s'", t)
	}

	if name == "" {
		v.SetGenerateName(k.prefix)
	}

	labels := v.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels["nerd-app"] = "cli"
	v.SetLabels(labels)

	k.logs.Debugf("creating %s '%s' with labels '%v': %s", t, v.GetName(), labels, ctx)
	err = c.Post().
		Resource(string(t)).
		Body(vv).
		Context(ctx).
		Do().
		Into(vv)

	if err != nil {
		return k.tagError(err)
	}

	v.SetName(k.removePrefix(v.GetName())) //normalize back to unprefixed resource name
	return nil
}

//UpdateResource will use the kube RESTClient to update a resource while using the context, adding the
//Nerd prefix and handling errors specific to our domain.
func (k *Visor) UpdateResource(ctx context.Context, t ResourceType, v ManagedNames, name string) (err error) {
	vv, ok := v.(runtime.Object)
	if !ok {
		return errors.Errorf("provided value was not castable to runtime.Object")
	}

	var c rest.Interface
	switch t {
	case ResourceTypeJobs:
		c = k.api.BatchV1().RESTClient()
	case ResourceTypePods, ResourceTypeSecrets, ResourceTypePersistentVolumeClaims:
		c = k.api.CoreV1().RESTClient()
	case ResourceTypeDatasets:
		c = k.crd.NerdalizeV1().RESTClient()
	default:
		return errors.Errorf("unknown Kubernetes resource type provided for update: '%s'", t)
	}

	name = k.applyPrefix(name)
	v.SetName(name)

	k.logs.Debugf("updating %s '%s' in namespace '%s': %s", t, v.GetName(), k.ns, ctx)
	err = c.Put().
		Namespace(k.ns).
		Resource(string(t)).
		Body(vv).
		Name(name).
		Context(ctx).
		Do().
		Into(vv)
	if err != nil {
		return k.tagError(err)
	}

	v.SetName(k.removePrefix(v.GetName())) //normalize back to unprefixed resource name
	return nil
}

//ListResources will use the RESTClient to list resources while using the context and transparently
//filter resources managed by the CLI
func (k *Visor) ListResources(ctx context.Context, t ResourceType, v ListTranformer, lselector, fselector []string) (err error) {
	vv, ok := v.(runtime.Object)
	if !ok {
		return errors.Errorf("provided value was not castable to runtime.Object")
	}

	var c rest.Interface
	s := scheme.ParameterCodec
	switch t {
	case ResourceTypeJobs:
		c = k.api.BatchV1().RESTClient()
	case ResourceTypePods, ResourceTypeEvents, ResourceTypeQuota, ResourceTypeSecrets, ResourceTypePersistentVolumeClaims:
		c = k.api.CoreV1().RESTClient()
	case ResourceTypeDatasets:
		c = k.crd.NerdalizeV1().RESTClient()
		s = crdscheme.ParameterCodec
	default:
		return errors.Errorf("unknown Kubernetes resource type provided for listing: '%s'", t)
	}

	//events are not  created by us so cannot be selected by our nerd label
	if t != ResourceTypeEvents && t != ResourceTypeQuota {
		lselector = append(lselector, "nerd-app=cli")
	}

	//get all the resources matching the selectors
	err = c.Get().
		Namespace(k.ns).
		VersionedParams(&metav1.ListOptions{
			LabelSelector: strings.Join(lselector, ","),
			FieldSelector: strings.Join(fselector, ","),
		}, s).
		Resource(string(t)).
		Context(ctx).
		Do().
		Into(vv)

	if err != nil {
		return k.tagError(err)
	}

	//if we have zero resources the namespace might not exists
	if v.Len() == 0 {
		_, err = k.api.CoreV1().Namespaces().Get(k.ns, metav1.GetOptions{})
		if err != nil {
			return k.tagError(err)
		}
	}

	//when it comes to events we do not want to return non-cli events
	if t == ResourceTypeEvents {
		v.Transform(func(in ManagedNames) ManagedNames {
			if !k.hasPrefix(in.GetName()) {
				return nil
			}

			return in
		})
	}

	//transform each managed item to return with its name unprefixed
	v.Transform(func(in ManagedNames) ManagedNames {
		in.SetName(k.removePrefix(in.GetName()))
		return in
	})

	return nil
}

func (k *Visor) tagError(err error) error {
	if uerr, ok := err.(*url.Error); ok {
		if uerr.Err == context.DeadlineExceeded {
			return errDeadline{uerr}
		}

		if nerr, ok := uerr.Err.(*net.OpError); ok {
			return errNetwork{nerr}
		}
	}

	if serr, ok := err.(*kuberr.StatusError); ok {
		if kuberr.IsServiceUnavailable(err) {
			return errServiceUnavailable{err}
		}

		if kuberr.IsUnauthorized(err) {
			return errUnauthorized{err}
		}

		if kuberr.IsAlreadyExists(serr) {
			return errAlreadyExists{err}
		}

		if kuberr.IsNotFound(serr) {
			details := serr.ErrStatus.Details
			if details.Kind == "namespaces" {
				return errNamespaceNotExists{err}
			}

			return errNotExists{err}
		}

		if kuberr.IsInvalid(serr) {
			details := serr.ErrStatus.Details
			for _, cause := range details.Causes {
				if cause.Field == "metadata.name" {
					return errInvalidName{err}
				}
			}
		}
	}

	return errKubernetes{err} //generic kubernetes error
}
