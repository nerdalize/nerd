package cmd

import (
	"os"
	"path/filepath"
	"time"
	"fmt"
	"strings"

	"github.com/nerdalize/nerd/nerd"

	"github.com/go-playground/validator"
	homedir "github.com/mitchellh/go-homedir"
	crd "github.com/nerdalize/nerd/crd/pkg/client/clientset/versioned"
	"github.com/nerdalize/nerd/pkg/populator"
	"github.com/nerdalize/nerd/pkg/transfer"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//TransferOpts hold CLI options for configuring data transfer
type TransferOpts struct {
	AWSS3Bucket        string `long:"aws-s3-bucket" description:"AWS S3 Bucket name that will be used for dataset storage" default:"nlz-datasets-dev"`
	AWSRegion          string `long:"aws-region" description:"AWS region used for dataset storage"`
	AWSAccessKeyID     string `long:"aws-access-key-id" description:"AWS access key used for auth with the storage backend"`
	AWSSecretAccessKey string `long:"aws-secret-access-key" description:"AWS secret key for auth with the storage backend"`
	AWSSessionToken    string `long:"aws-session-token" description:"AWS temporary auth token for the storage backend"`
}

//Transfer creates an concrete transfer service using the configuration
func (opts TransferOpts) Transfer() (trans transfer.Transfer, err error) {
	s3cfg := &transfer.S3Conf{
		Bucket:       opts.AWSS3Bucket,
		Region:       opts.AWSRegion,
		AccessKey:    opts.AWSAccessKeyID,
		SecretKey:    opts.AWSSecretAccessKey,
		SessionToken: opts.AWSSessionToken,
	}

	trans, err = transfer.NewS3(s3cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create s3 uploader")
	}

	return trans, nil
}

//KubeOpts can be used to create a Kubernetes service
type KubeOpts struct {
	KubeConfig string        `long:"kube-config" description:"file at which Nerd will look for Kubernetes credentials" env:"KUBECONFIG" default-mask:"~/.kube/conf"`
	Timeout    time.Duration `long:"timeout" description:"duration for which Nerd will wait for Kubernetes" default-mask:"10s" default:"10s" required:"true"`
}

//Deps exposes dependencies
type Deps struct {
	val  svc.Validator
	kube kubernetes.Interface
	crd  crd.Interface
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
		if os.IsNotExist(err) {
			return nil, ErrNotLoggedIn
		}
		return nil, errors.Wrap(err, "failed to build Kubernetes config from provided kube config path")
	}

	d := &Deps{
		logs: logs,
	}

	d.crd, err = crd.NewForConfig(kcfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Kubernetes configuration")
	}

	d.kube, err = kubernetes.NewForConfig(kcfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Kubernetes configuration")
	}

	if !populator.Context(kopts.KubeConfig) {
		return nil, nerd.ErrProjectIDNotSet
	}

	d.ns, err = populator.Namespace(kopts.KubeConfig)
	if err != nil || d.ns == "" {
		return nil, nerd.ErrProjectIDNotSet
	}

	val := validator.New()
	val.RegisterValidation("is-abs-path", validateAbsPath)
	d.val = val

	return d, nil
}

//Derived from https://github.com/golang/go/blob/1106512db54fc2736c7a9a67dd553fc9e1fca742/src/path/filepath/path_unix.go#L12
func validateAbsPath(fl validator.FieldLevel) bool {
	return strings.HasPrefix(fl.Field().String(), "/")
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

//Crd provides the custom resource definition API
func (deps *Deps) Crd() crd.Interface {
	return deps.crd
}
