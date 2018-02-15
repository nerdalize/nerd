package svc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator"
	homedir "github.com/mitchellh/go-homedir"
	crd "github.com/nerdalize/nerd/crd/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

//Validator describes the validation dependency we require
type Validator interface {
	StructCtx(ctx context.Context, s interface{}) (err error)
}

//Logger describes the logging dependency the services require
type Logger interface {
	Debugf(format string, args ...interface{})
}

//DI provides dependencies for our services
type DI interface {
	Kube() kubernetes.Interface
	Crd() crd.Interface
	Validator() Validator
	Logger() Logger
	Namespace() string
}

//ErrMinikubeOnly is returned when a temp di is created on something thats not minikube
var ErrMinikubeOnly = errors.New("temp DI can only be created on Minikube")

//TempDI returns a temporary DI for the Kube service that sets up
//a temporary namespace which can be deleted using clean. Mostly
//usefull for testing purposes. If name is empty a 16 byte random
//one will be generated
func TempDI(name string) (di DI, clean func(), err error) {
	hdir, err := homedir.Dir()
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to determine homedir")
	}

	kcfg, err := clientcmd.BuildConfigFromFlags("", filepath.Join(hdir, ".kube", "config"))
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to build config form kube config")
	}

	if !strings.Contains(fmt.Sprintf("%#v", kcfg), "minikube") {
		return nil, nil, ErrMinikubeOnly
	}

	tdi := &tmpDI{
		logs: logrus.New(),
		val:  validator.New(),
	}
	tdi.kube, err = kubernetes.NewForConfig(kcfg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create kube client from config")
	}

	tdi.crd, err = crd.NewForConfig(kcfg)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create CRD client from config")
	}

	if name == "" {
		d := make([]byte, 16)
		_, err = rand.Read(d)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed to read random bytes")
		}

		name = hex.EncodeToString(d)
	}

	ns, err := tdi.kube.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{GenerateName: name},
	})

	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create temporary namespace")
	}

	tdi.ns = ns.GetName()
	return tdi, func() {
		err := tdi.kube.CoreV1().Namespaces().Delete(ns.Name, nil)
		if err != nil {
			panic(err)
		}
	}, nil
}

type tmpDI struct {
	kube kubernetes.Interface
	crd  crd.Interface
	val  Validator
	logs Logger
	ns   string
}

func (di *tmpDI) Kube() kubernetes.Interface { return di.kube }

func (di *tmpDI) Validator() Validator { return di.val }

func (di *tmpDI) Logger() Logger { return di.logs }

func (di *tmpDI) Namespace() string { return di.ns }

func (di *tmpDI) Crd() crd.Interface { return di.crd }
