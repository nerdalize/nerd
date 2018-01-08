package cmd

import (
	"path/filepath"
	"time"

	"github.com/nerdalize/nerd/nerd"

	"github.com/go-playground/validator"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/nerdalize/nerd/pkg/populator"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//KubeOpts can be used to create a Kubernetes service
type KubeOpts struct {
	KubeConfig string        `long:"kube-config" description:"file at which Nerd will look for Kubernetes credentials" env:"KUBECONFIG" default-mask:"~/.kube/conf"`
	Timeout    time.Duration `long:"timeout" description:"duration for which Nerd will wait for Kubernetes" default-mask:"10s" default:"10s" required:"true"`
}

//Deps exposes dependencies
type Deps struct {
	val  svc.Validator
	kube kubernetes.Interface
	logs svc.Logger
	ns   string
}

//NewDeps uses options to setup dependencies
func NewDeps(logs svc.Logger, kopts KubeOpts) (*Deps, error) {
	if kopts.KubeConfig == "" {
		hdir, err := homedir.Dir()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get home directory")
		}

		kopts.KubeConfig = filepath.Join(hdir, ".kube", "config")
	}

	kcfg, err := clientcmd.BuildConfigFromFlags("", kopts.KubeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build Kubernetes config from provided kube config path")
	}

	d := &Deps{
		logs: logs,
	}

	d.kube, err = kubernetes.NewForConfig(kcfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Kubernetes configuration")
	}

	if !populator.Context(kopts.KubeConfig) {
		return nil, nerd.ErrProjectIDNotSet
	}

	d.ns, err = populator.Namespace(kopts.KubeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get namespace from Kubernetes configuration")
	}

	d.val = validator.New()
	return d, nil
}

//Kube provides the kubernetes dependency
func (deps *Deps) Kube() kubernetes.Interface {
	return deps.kube
}

//Validator provides the Validator dependency
func (deps *Deps) Validator() svc.Validator {
	return deps.val
}

//Logger provides the Logger dependency
func (deps *Deps) Logger() svc.Logger {
	return deps.logs
}

//Namespace provides the namespace dependency
func (deps *Deps) Namespace() string {
	return deps.ns
}
