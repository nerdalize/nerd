//main holds the flex volume executable, compiled separately
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/go-playground/validator"
	"github.com/joho/godotenv"
	crd "github.com/nerdalize/nerd/crd/pkg/client/clientset/versioned"
	"github.com/nerdalize/nerd/pkg/transfer"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	certutil "k8s.io/client-go/util/cert"
)

//DevNullLogger is used to disable kube visor logging
type DevNullLogger struct{}

//Debugf implementation
func (l *DevNullLogger) Debugf(format string, args ...interface{}) {}

//Operation can be provided to the flex volume
type Operation string

const (
	//OperationInit is called when the flex volume needs to set itself up
	OperationInit = "init"

	//OperationMount is called when a volume needs to be mounted
	OperationMount = "mount"

	//OperationUnmount is called when the volume needs to be unmounted
	OperationUnmount = "unmount"
)

//Status describes a flex volume status
type Status string

const (
	//StatusSuccess is returned when the flex volume has been successfull
	StatusSuccess = "Success"
	//StatusFailure is returned when the flex volume has failed
	StatusFailure = "Failure"
	//StatusNotSupported is returned when a operation is supported
	StatusNotSupported = "Not supported"
)

//Output is returned by the flex volume implementation
type Output struct {
	Status       Status       `json:"status"`
	Message      string       `json:"message"`
	Capabilities Capabilities `json:"capabilities"`
}

//MountOptions is specified whenever Kubernetes calls the mount, comes with
//the following keys: kubernetes.io/fsType, kubernetes.io/pod.name, kubernetes.io/pod.namespace
//kubernetes.io/pod.uid, kubernetes.io/pvOrVolumeName, kubernetes.io/readwrite, kubernetes.io/serviceAccount.name
type MountOptions struct {
	InputDataset  string `json:"input/dataset"`
	OutputDataset string `json:"output/dataset"`
	Namespace     string `json:"kubernetes.io/pod.namespace"`
}

//Capabilities of the flex volume
type Capabilities struct {
	Attach bool `json:"attach"`
}

//VolumeDriver can be implemented to facilitate the creation of pod volumes
type VolumeDriver interface {
	Init() (Capabilities, error)
	Mount(mountPath string, opts MountOptions) error
	Unmount(mountPath string) error
}

//DatasetVolumes is a volume implementations that works with Nerdalize Datasets
type DatasetVolumes struct{}

type datasetOpts struct {
	Namespace     string
	InputDataset  string
	OutputDataset string
}

func (volp *DatasetVolumes) writeDatasetOpts(mountPath string, opts MountOptions) (*datasetOpts, error) {
	dsopts := &datasetOpts{Namespace: opts.Namespace}
	if dsopts.Namespace == "" {
		return nil, errors.New("pod namespace was not configured for flex volume")
	}

	dsopts.InputDataset = opts.InputDataset
	dsopts.OutputDataset = opts.OutputDataset

	path := filepath.Join(mountPath, "..", filepath.Base(mountPath)+".json")
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %v", err)
	}

	defer f.Close()
	enc := json.NewEncoder(f)
	err = enc.Encode(dsopts)
	if err != nil {
		return nil, fmt.Errorf("failed to encode options: %v", err)
	}

	return dsopts, nil
}

func (volp *DatasetVolumes) readDatasetOpts(mountPath string) (*datasetOpts, error) {
	path := filepath.Join(mountPath, "..", filepath.Base(mountPath)+".json")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	defer f.Close()
	dsopts := &datasetOpts{}

	dec := json.NewDecoder(f)
	err = dec.Decode(dsopts)
	if err != nil {
		return nil, fmt.Errorf("failed to decode options")
	}

	return dsopts, nil
}

//Init the flex volume
func (volp *DatasetVolumes) Init() (Capabilities, error) {
	return Capabilities{Attach: false}, nil
}

//Mount the flex voume, path: '/var/lib/kubelet/pods/c911e5f7-0392-11e8-8237-32f9813bbd5a/volumes/foo~cifs/input', opts: &main.MountOptions{FSType:"", PodName:"imagemagick", PodNamespace:"default", PodUID:"c911e5f7-0392-11e8-8237-32f9813bbd5a", PVOrVolumeName:"input", ReadWrite:"rw", ServiceAccountName:"default"}
func (volp *DatasetVolumes) Mount(mountPath string, opts MountOptions) error {
	dsopts, err := volp.writeDatasetOpts(mountPath, opts)
	if err != nil {
		return fmt.Errorf("failed to write volume database: %v", err)
	}

	if dsopts.InputDataset == "" {
		return nil //no input for volume
	}

	//
	// EXPERIMENTAL
	//

	di, err := NewDeps(dsopts.Namespace)
	if err != nil {
		return errors.Wrapf(err, "failed to setup dependencies")
	}

	kube := svc.NewKube(di)
	out, err := kube.GetDataset(context.TODO(), &svc.GetDatasetInput{Name: dsopts.InputDataset})
	if err != nil {
		//@TODO if the dataset doesn't exist it might be deleted, error more gracefully
		//@TODO throw a warning or error to the user, allow retry, how many times?
		return errors.Wrap(err, "failed to get dataset")
	}

	var trans transfer.Transfer
	if trans, err = transfer.NewS3(&transfer.S3Conf{
		Bucket: out.Bucket,
	}); err != nil {
		return err
	}

	ref := &transfer.Ref{
		Bucket: out.Bucket,
		Key:    out.Key,
	}

	//
	// EXPERIMENTAL
	//

	//@TODO when this fails flex volume retry mechanism will never succeed because the directory is not empty
	err = trans.Download(context.Background(), ref, mountPath)
	if err != nil {
		return errors.Wrapf(err, "failed to download to '%s'", mountPath)
	}

	return nil
}

//Unmount the flex voume
func (volp *DatasetVolumes) Unmount(mountPath string) (err error) {
	var dsopts *datasetOpts
	dsopts, err = volp.readDatasetOpts(mountPath)
	if err != nil {
		return fmt.Errorf("failed to read volume database: %v", err)
	}

	defer func() {
		if err == nil { //if there was no error during upload remove all data
			err = os.RemoveAll(mountPath)
		}
	}()

	if dsopts.OutputDataset == "" {
		return nil //no output dataset, do nothing with the volume data
	}

	//
	// EXPERIMENTAL
	//

	di, err := NewDeps(dsopts.Namespace)
	if err != nil {
		return errors.Wrapf(err, "failed to setup dependencies")
	}

	kube := svc.NewKube(di)
	out, err := kube.GetDataset(context.TODO(), &svc.GetDatasetInput{Name: dsopts.OutputDataset})
	if err != nil {
		//@TODO if the dataset doesn't exist it might be deleted, error more gracefully
		return errors.Wrap(err, "failed to get dataset")
	}

	var trans transfer.Transfer
	if trans, err = transfer.NewS3(&transfer.S3Conf{
		Bucket: out.Bucket,
	}); err != nil {
		return err
	}

	ref := &transfer.Ref{
		Bucket: out.Bucket,
		Key:    out.Key,
	}

	//
	// EXPERIMENTAL
	//

	_, err = trans.Upload(context.Background(), ref, mountPath)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: nerd-flex-volume [init|mount|unmount]")
	}

	//create the volume provider
	var volp VolumeDriver
	volp = &DatasetVolumes{}

	//setup default output data
	var err error
	output := Output{
		Status:  StatusNotSupported,
		Message: fmt.Sprintf("operation '%s' is unsupported", os.Args[1]),
	}

	//pass operations to the volume provider
	switch os.Args[1] {
	case OperationInit:
		output.Status = StatusSuccess
		output.Capabilities, err = volp.Init()

	case OperationMount:
		output.Status = StatusSuccess
		if len(os.Args) < 4 {
			err = fmt.Errorf("expected at least 4 arguments for mount, got: %#v", os.Args)
		} else {
			opts := MountOptions{}
			err = json.Unmarshal([]byte(os.Args[3]), &opts)
			if err == nil {
				err = volp.Mount(os.Args[2], opts)
			}
		}

	case OperationUnmount:
		output.Status = StatusSuccess
		if len(os.Args) < 3 {
			err = fmt.Errorf("expected at least 3 arguments for unmount, got: %#v", os.Args)
		} else {
			err = volp.Unmount(os.Args[2])
		}
	}

	//if any operations returned an error, mark as failure
	if err != nil {
		output.Status = StatusFailure
		output.Message = err.Error()
	}

	//encode the output
	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(output)
	if err != nil {
		log.Fatalf("failed to encode output: %v", err)
	}
}

//CreateKubernetesConfig will read a envionment file and service account
//specifically setup to provide a connection from the host to the API server
func CreateKubernetesConfig() (*rest.Config, error) {
	exep, err := os.Executable()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load executable path")
	}

	exedir := filepath.Join(filepath.Dir(exep))

	//read environment from .env file
	err = godotenv.Load(filepath.Join(exedir, "flex.env"))
	if err != nil {
		return nil, errors.Wrap(err, "failed to load flex environment")
	}

	//read token file from service account
	token, err := ioutil.ReadFile(filepath.Join(exedir, "serviceaccount", v1.ServiceAccountTokenKey))
	if err != nil {
		return nil, errors.Wrap(err, "failed to read service account token key")
	}

	//read CA config from service account
	tlsClientConfig := rest.TLSClientConfig{}
	rootCAFile := filepath.Join(exedir, "serviceaccount", v1.ServiceAccountRootCAKey)
	if _, err = certutil.NewPool(rootCAFile); err != nil {
		return nil, errors.Wrap(err, "failed to load service account CA files")
	}

	tlsClientConfig.CAFile = rootCAFile

	//read kubernetes api host and port from (imported) evironment
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		return nil, errors.Errorf("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined")
	}

	//create rest config
	config := &rest.Config{
		Host:            "https://" + net.JoinHostPort(host, port),
		BearerToken:     string(token),
		TLSClientConfig: tlsClientConfig,
	}

	return config, nil
}

//Deps holds flex volume dependencies to setup
//our kubernetes service
type Deps struct {
	val  svc.Validator
	kube kubernetes.Interface
	crd  crd.Interface
	logs svc.Logger
	ns   string
}

//NewDeps sets up the Kubernetes service dependencies specifically for
//the flex volume client
func NewDeps(namespace string) (d *Deps, err error) {
	d = &Deps{
		ns:   namespace,
		val:  validator.New(),
		logs: &DevNullLogger{},
	}

	kcfg, err := CreateKubernetesConfig()
	if err != nil {
		return d, errors.Wrap(err, "failed to setup Kubernetes connection")
	}

	d.crd, err = crd.NewForConfig(kcfg)
	if err != nil {
		return d, errors.Wrap(err, "failed to create Kubernetes CRD interface")
	}

	d.kube, err = kubernetes.NewForConfig(kcfg)
	if err != nil {
		return d, errors.Wrap(err, "failed to create Kubernetes configuration")
	}

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

//Crd returns the custom resource depenition interface
func (deps *Deps) Crd() crd.Interface {
	return deps.crd
}
