package svc

import (
	"context"
	"net/url"
	"strings"

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

//KubeNameable allows our create abstraction to set names prior to creation
type KubeNameable interface {
	GetName() string
	SetName(name string)
	SetGenerateName(name string)
}

//Kube interacts with the kubernetes backend
type Kube struct {
	prefix string
	ns     string
	api    kubernetes.Interface
	val    Validator
}

//NewKube will setup the Kubernetes service
func NewKube(di DI, ns string) (k *Kube) {
	k = &Kube{
		prefix: "n.e.r.d-",
		ns:     ns,
		api:    di.Kube(),
		val:    di.Validator(),
	}

	return k
}

//CreateResource will use the kube RESTClient to create a resource while using the context, adding the
//Nerd prefix and handling errors specific to our domain.
func (k *Kube) createResource(ctx context.Context, t KubeResourceType, v KubeNameable, name string) (err error) {

	vv, ok := v.(runtime.Object)
	if !ok {
		return errors.Errorf("provided value was not castable to runtime.Object")
	}

	genfix := "x-"
	var c rest.Interface
	switch t {
	case KubeResourceTypeJobs:
		c = k.api.BatchV1().RESTClient()
		genfix = "j-"

	default:
		return errors.Errorf("unknown Kubernetes resource type provided: '%s'", t)
	}

	if name != "" {
		v.SetName(k.prefix + name)
	} else {
		v.SetGenerateName(k.prefix + genfix)
	}

	err = c.Post().
		Namespace(k.ns).
		Resource(string(t)).
		Body(vv).
		Context(ctx).
		Do().
		Into(vv)

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

	v.SetName(strings.TrimPrefix(v.GetName(), k.prefix)) //normalize back to unprefixed resource name
	return nil
}
