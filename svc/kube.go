package svc

import (
	"context"
	"net/url"

	"github.com/pkg/errors"

	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

//KubeResourceType is a type of Kubernetes resource
type KubeResourceType string

var (
	//KubeResourceTypeJobs is used for job management
	KubeResourceTypeJobs = KubeResourceType("jobs")
)

//Kube interacts with the kubernetes backend
type Kube struct {
	ns  string
	api kubernetes.Interface
	val Validator
}

//NewKube will setup the Kubernetes service
func NewKube(di DI, ns string) (k *Kube) {
	k = &Kube{
		ns:  ns,
		api: di.Kube(),
		val: di.Validator(),
	}

	return k
}

//CreateResource will use the kube RESTClient to create a resource while using the context, adding the
//nerd prefix and handling errors specific to our domain.
func (k *Kube) CreateResource(ctx context.Context, t KubeResourceType, v runtime.Object) (err error) {
	var c rest.Interface
	switch t {
	case KubeResourceTypeJobs:
		c = k.api.BatchV1().RESTClient()
	default:
		return errors.Errorf("unknown Kubernetes resource type provided: '%s'", t)
	}

	err = c.Post().
		Namespace(k.ns).
		Resource(string(t)).
		Body(v).
		Context(ctx).
		Do().
		Into(v)

	if err != nil {
		if uerr, ok := err.(*url.Error); ok && uerr.Err == context.DeadlineExceeded {
			return errDeadline{uerr}
		}

		if serr, ok := err.(*kuberr.StatusError); ok {
			if kuberr.IsAlreadyExists(serr) {
				return errAlreadyExists{err}
			}
		}

		return errKubernetes{err} //generic kubernetes error
	}

	return nil
}
