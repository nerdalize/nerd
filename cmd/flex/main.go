//main holds the flex volume executable, compiled separately
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	crd "github.com/nerdalize/nerd/crd/pkg/client/clientset/versioned"
	transfer "github.com/nerdalize/nerd/pkg/transfer"
	transferarchiver "github.com/nerdalize/nerd/pkg/transfer/archiver"
	"github.com/nerdalize/nerd/svc"

	"github.com/go-playground/validator"
	"github.com/hashicorp/go-multierror"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	certutil "k8s.io/client-go/util/cert"
)

//DevNullLogger is used to disable kube visor logging
type DevNullLogger struct{}

//Debugf implementation
func (l *DevNullLogger) Debugf(format string, args ...interface{}) {}

//Operation is an action that can be performed with the flex volume.
type Operation string

const (
	//OperationInit is called when the flex volume needs to set itself up
	OperationInit = "init"

	//OperationMount is called when a volume needs to be mounted
	OperationMount = "mount"

	//OperationUnmount is called when the volume needs to be unmounted
	OperationUnmount = "unmount"
)

//Status describes the result of a flex volume action.
type Status string

const (
	//StatusSuccess is returned when the flex volume has been successfull
	StatusSuccess = "Success"
	//StatusFailure is returned when the flex volume has failed
	StatusFailure = "Failure"
	//StatusNotSupported is returned when a operation is supported
	StatusNotSupported = "Not supported"
)

//FileSystem can be used to specify a type of file system in a file.
type FileSystem string

const (
	//FileSystemExt4 is the standard, supported everywhere
	FileSystemExt4 FileSystem = "ext4"
)

//WriteSpace is the amount of space available for writing data.
//@TODO: Should be based on dataset size or customer details?
const WriteSpace = 2 * 1024 * 1024 * 1024

//DirectoryPermissions are the permissions for directories created as part of flexvolume operation.
//@TODO: Spend more time checking if they make sense and are secure
const DirectoryPermissions = os.FileMode(0522)

//Relative paths used for flexvolume data
const (
	RelPathInput         = "input"
	RelPathFSInFile      = "volume"
	RelPathFSInFileMount = "mount"
	RelPathOptions       = "json"
	LogFile              = "/var/lib/kubelet/flex.logs"
)

//Output is returned by the flex volume implementation.
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

//Capabilities represents the supported features of a flex volume.
type Capabilities struct {
	Attach bool `json:"attach"`
}

//VolumeDriver can be implemented to facilitate the creation of pod volumes.
type VolumeDriver interface {
	Init() (Capabilities, error)
	Mount(mountPath string, opts MountOptions) error
	Unmount(mountPath string) error
}

//DatasetVolumes is a volume implementation that works with Nerdalize Datasets.
type DatasetVolumes struct{}

//datasetOpts describes any input and output for a volume.
type datasetOpts struct {
	Namespace     string
	InputDataset  string
	OutputDataset string
}

//writeDatasetOpts writes dataset options to a JSON file.
func (volp *DatasetVolumes) writeDatasetOpts(path string, opts MountOptions) (*datasetOpts, error) {
	dsopts := &datasetOpts{Namespace: opts.Namespace}
	if dsopts.Namespace == "" {
		return nil, errors.New("pod namespace was not configured for flex volume")
	}

	dsopts.InputDataset = opts.InputDataset
	dsopts.OutputDataset = opts.OutputDataset

	f, err := os.Create(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create metadata file")
	}

	defer f.Close()

	enc := json.NewEncoder(f)
	err = enc.Encode(dsopts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encode metadata")
	}

	return dsopts, nil
}

//readDatasetOpts reads dataset options from a JSON file.
func (volp *DatasetVolumes) readDatasetOpts(path string) (*datasetOpts, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open metadata file")
	}

	defer f.Close()
	dsopts := &datasetOpts{}

	dec := json.NewDecoder(f)
	err = dec.Decode(dsopts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode metadata")
	}

	return dsopts, nil
}

//deleteDatasetOpts deletes a JSON file containing dataset options.
func (volp *DatasetVolumes) deleteDatasetOpts(path string) error {
	err := os.Remove(path)
	return errors.Wrap(err, "failed to delete metadata file")
}

//createFSInFile creates a file with a file system inside of it that can be mounted.
func (volp *DatasetVolumes) createFSInFile(path string, filesystem FileSystem, size int64) error {
	//Create file with room to contain writable file system
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "failed to create file system file")
	}

	err = f.Truncate(size)
	if err != nil {
		return errors.Wrap(err, "failed to allocate file system size")
	}

	//Build file system within
	cmd := exec.Command("mkfs", "-t", string(filesystem), path)
	buf := bytes.NewBuffer(nil)
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(errors.New(strings.TrimSpace(buf.String())), "failed to execute mkfs command")
	}

	return nil
}

//destroyFSInFile cleans up a file system in file.
func (volp *DatasetVolumes) destroyFSInFile(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		err = errors.Wrap(err, "failed to delete fs-in-file file")
	}

	return err
}

func (volp *DatasetVolumes) transferManager(kube *svc.Kube) (mgr transfer.Manager, err error) {
	if mgr, err = transfer.NewKubeManager(kube); err != nil {
		return nil, errors.Wrap(err, "failed to setup transfer manager")
	}

	return mgr, nil
}

//provisionInput makes the specified input available at given path (input may be nil).
func (volp *DatasetVolumes) provisionInput(path, namespace, dataset string) error {
	//Create directory at path in case it doesn't exist yet
	err := os.MkdirAll(path, DirectoryPermissions)
	if err != nil {
		return errors.Wrap(err, "failed to create input directory")
	}

	//Abort if there is nothing to download to it
	if dataset == "" {
		return nil
	}

	di, err := NewDeps(namespace)
	if err != nil {
		return errors.Wrap(err, "failed to setup dependencies")
	}

	mgr, err := volp.transferManager(svc.NewKube(di))
	if err != nil {
		return errors.Wrap(err, "failed to setup transfer manager")
	}

	ctx := context.TODO() //@TODO decide on a deadline for this

	h, err := mgr.Open(ctx, dataset)
	if err != nil {
		return errors.Wrap(err, "failed to open dataset")
	}

	defer h.Close()
	err = h.Pull(ctx, path, transfer.NewDiscardReporter())
	if err != nil {
		return errors.Wrap(err, "failed to download dataset")
	}

	return nil
}

//destroyInput cleans up a folder with input data.
func (volp *DatasetVolumes) destroyInput(path string) error {
	return errors.Wrap(os.RemoveAll(path), "failed to destroy input directory")
}

//mountFSInFile mounts an FS-in-file at the specified path.
func (volp *DatasetVolumes) mountFSInFile(volumePath string, mountPath string) error {
	//Create mount point
	err := os.Mkdir(mountPath, DirectoryPermissions)
	if err != nil {
		return errors.Wrap(err, "failed to create mount directory")
	}

	//Mount file system
	cmd := exec.Command("mount", volumePath, mountPath)
	buf := bytes.NewBuffer(nil)
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(errors.New(strings.TrimSpace(buf.String())), "failed to execute mount command")
	}

	return nil
}

//unmountFSInFile unmounts an FS-in-file and deletes the mount path.
func (volp *DatasetVolumes) unmountFSInFile(mountPath string) error {
	//Unmount
	cmd := exec.Command("umount", mountPath)
	buf := bytes.NewBuffer(nil)
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(errors.New(strings.TrimSpace(buf.String())), "failed to unmount fs-in-file")
	}

	//Delete mount path
	err = os.RemoveAll(mountPath)
	if err != nil {
		return errors.Wrap(err, "failed to delete fs-in-file mount point")
	}

	return nil
}

//mountOverlayFS mounts an OverlayFS with the given directories (upperDir and workDir may be auto-created).
func (volp *DatasetVolumes) mountOverlayFS(upperDir string, workDir string, lowerDir string, mountPath string) error {
	//Create directories in case they don't exist yet
	errs := []error{
		os.MkdirAll(upperDir, DirectoryPermissions),
		os.MkdirAll(workDir, DirectoryPermissions),
	}

	for _, err := range errs {
		if err != nil {
			return errors.Wrap(err, "failed to create directories")
		}
	}

	//Mount OverlayFS
	overlayArgs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir)

	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", overlayArgs, mountPath)
	buf := bytes.NewBuffer(nil)
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(errors.New(strings.TrimSpace(buf.String())), "failed to execute mount command")
	}

	return nil
}

//unmountOverlayFS unmounts an OverlayFS with the given directories (upperDir and workDir will be deleted).
func (volp *DatasetVolumes) unmountOverlayFS(upperDir string, workDir string, mountPath string) error {
	//Unmount OverlayFS
	cmd := exec.Command("umount", mountPath)
	buf := bytes.NewBuffer(nil)
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(errors.New(strings.TrimSpace(buf.String())), "failed to unmount overlayfs")
	}

	//Delete directories
	errs := []error{
		os.RemoveAll(upperDir),
		os.RemoveAll(workDir),
	}

	for _, err := range errs {
		if err != nil {
			return errors.Wrap(err, "failed to delete directories")
		}
	}

	return nil
}

//handleOutput uploads any output in the specified directory.
func (volp *DatasetVolumes) handleOutput(path, namespace, dataset string) error {
	// Nothing to do
	if dataset == "" {
		return nil
	}

	di, err := NewDeps(namespace)
	if err != nil {
		return errors.Wrap(err, "failed to setup dependencies")
	}

	mgr, err := volp.transferManager(svc.NewKube(di))
	if err != nil {
		return errors.Wrap(err, "failed to setup transfer manager")
	}

	ctx := context.TODO() //@TODO decide on a deadline for this

	h, err := mgr.Open(ctx, dataset)

	// If the user has deleted the dataset, then there is nothing to do
	if err != nil {
		log.Printf("warning, output dataset no longer exists: %v\n", err)
		return nil
	}

	defer h.Close()
	err = h.Push(ctx, path, transfer.NewDiscardReporter())

	// The output dataset being empty is a non-fatal unmount error
	if err != nil && strings.Contains(err.Error(), transferarchiver.ErrEmptyDirectory.Error()) {
		log.Printf("warning, output dataset is empty: %v\n", err)
		return nil
	}

	if err != nil {
		return errors.Wrap(err, "failed to transfer dataset")
	}

	return nil
}

// fetchAllowedSpace is a temporary solution so we can give more space to specific users on the public cluster
func (volp *DatasetVolumes) fetchAllowedSpace(path, namespace string) (space int64, err error) {
	// TODO WriteSpace should be used only if there is no label "flex-volume-size" in the namespace quota labels.
	space = WriteSpace

	di, err := NewDeps(namespace)
	if err != nil {
		return space, errors.Wrap(err, "failed to setup dependencies")
	}

	kube := svc.NewKube(di)
	quotas, err := kube.ListQuotas(context.TODO(), &svc.ListQuotasInput{})
	if err != nil {
		return space, err
	}
	if quotas != nil && len(quotas.Items) == 0 {
		return space, err
	}
	if quotas.Items[0].Labels["flex-volume-size"] != "" {
		space, err = strconv.ParseInt(quotas.Items[0].Labels["flex-volume-size"], 10, 64)
		if err != nil {
			space = WriteSpace
		}
	}
	return space, err
}

//getPath returns a path above the mountPath and unique to the dataset name.
func (volp *DatasetVolumes) getPath(mountPath string, name string) string {
	return filepath.Join(mountPath, "..", filepath.Base(mountPath)+"."+name)
}

//cleanDirectory deletes the contents of a directory, but not the directory itself.
func (volp *DatasetVolumes) cleanDirectory(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()

	names, err := dir.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(path, name))
		if err != nil {
			return err
		}
	}

	return nil
}

//Init the flex volume.
func (volp *DatasetVolumes) Init() (Capabilities, error) {
	return Capabilities{Attach: false}, nil
}

//Mount the flex volume, path: '/var/lib/kubelet/pods/c911e5f7-0392-11e8-8237-32f9813bbd5a/volumes/foo~cifs/input', opts: &main.MountOptions{FSType:"", PodName:"imagemagick", PodNamespace:"default", PodUID:"c911e5f7-0392-11e8-8237-32f9813bbd5a", PVOrVolumeName:"input", ReadWrite:"rw", ServiceAccountName:"default"}
func (volp *DatasetVolumes) Mount(kubeMountPath string, opts MountOptions) (err error) {
	//Store dataset options
	dsopts, err := volp.writeDatasetOpts(volp.getPath(kubeMountPath, RelPathOptions), opts)

	defer func() {
		if err != nil {
			volp.deleteDatasetOpts(volp.getPath(kubeMountPath, RelPathOptions))
		}
	}()

	if err != nil {
		return errors.Wrap(err, "failed to write volume database")
	}

	//+TODO create kube here and inject it in provisionInput and fetchAllowedSpace
	//TODO create a context with a deadline
	//Set up input
	err = volp.provisionInput(volp.getPath(kubeMountPath, RelPathInput), dsopts.Namespace, dsopts.InputDataset)

	defer func() {
		if err != nil {
			volp.destroyInput(volp.getPath(kubeMountPath, RelPathInput))
		}
	}()

	if err != nil {
		return errors.Wrap(err, "failed to provision input")
	}

	writeSpace, err := volp.fetchAllowedSpace(kubeMountPath, dsopts.Namespace)
	if err != nil {
		return errors.Wrap(err, "failed to fetch allowed space")
	}
	//Create volume to contain pod writes
	err = volp.createFSInFile(volp.getPath(kubeMountPath, RelPathFSInFile), FileSystemExt4, writeSpace)

	defer func() {
		if err != nil {
			volp.destroyFSInFile(volp.getPath(kubeMountPath, RelPathFSInFile))
		}
	}()

	if err != nil {
		return errors.Wrap(err, "failed to create file system in a file")
	}

	//Mount the file system
	err = volp.mountFSInFile(
		volp.getPath(kubeMountPath, RelPathFSInFile),
		volp.getPath(kubeMountPath, RelPathFSInFileMount),
	)

	defer func() {
		if err != nil {
			volp.unmountFSInFile(volp.getPath(kubeMountPath, RelPathFSInFileMount))
		}
	}()

	if err != nil {
		return errors.Wrap(err, "failed to mount file system in a file")
	}

	//Set up overlay file system using input and writable fs-in-file
	err = volp.mountOverlayFS(
		filepath.Join(volp.getPath(kubeMountPath, RelPathFSInFileMount), "upper"),
		filepath.Join(volp.getPath(kubeMountPath, RelPathFSInFileMount), "work"),
		volp.getPath(kubeMountPath, RelPathInput),
		kubeMountPath,
	)

	defer func() {
		if err != nil {
			volp.unmountOverlayFS(
				filepath.Join(volp.getPath(kubeMountPath, RelPathFSInFileMount), "upper"),
				filepath.Join(volp.getPath(kubeMountPath, RelPathFSInFileMount), "work"),
				kubeMountPath,
			)
		}
	}()

	if err != nil {
		return errors.Wrap(err, "failed to mount overlayfs")
	}

	return nil
}

//Unmount the flex volume.
func (volp *DatasetVolumes) Unmount(kubeMountPath string) (err error) {
	// Upload any output
	var dsopts *datasetOpts
	dsopts, err = volp.readDatasetOpts(volp.getPath(kubeMountPath, RelPathOptions))
	if err != nil {
		log.Printf("warning: failed to read volume database at %s, assuming that volume has already been deleted: %v", kubeMountPath, err)
		return nil
	}

	err = volp.handleOutput(kubeMountPath, dsopts.Namespace, dsopts.OutputDataset)
	if err != nil {
		return errors.Wrap(err, "failed to upload output")
	}

	//Clean up (as much as possible)
	var result error

	err = errors.Wrap(
		volp.unmountOverlayFS(
			filepath.Join(volp.getPath(kubeMountPath, RelPathFSInFileMount), "upper"),
			filepath.Join(volp.getPath(kubeMountPath, RelPathFSInFileMount), "work"),
			kubeMountPath,
		),
		"failed to unmount overlayfs",
	)
	if err != nil {
		result = multierror.Append(result, err)
	}

	err = errors.Wrap(
		volp.unmountFSInFile(volp.getPath(kubeMountPath, RelPathFSInFileMount)),
		"failed to unmount file system in a file",
	)
	if err != nil {
		result = multierror.Append(result, err)
	}

	err = errors.Wrap(
		volp.destroyFSInFile(volp.getPath(kubeMountPath, RelPathFSInFile)),
		"failed to delete file system in a file",
	)
	if err != nil {
		result = multierror.Append(result, err)
	}

	err = errors.Wrap(
		volp.destroyInput(volp.getPath(kubeMountPath, RelPathInput)),
		"failed to delete input data",
	)
	if err != nil {
		result = multierror.Append(result, err)
	}

	err = errors.Wrap(
		volp.deleteDatasetOpts(volp.getPath(kubeMountPath, RelPathOptions)),
		"failed to delete dataset",
	)
	if err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func main() {
	var err error
	f, err := os.OpenFile(LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("flexvolume logs beginning")

	if len(os.Args) < 2 {
		fmt.Println("usage: nerd-flex-volume [init|mount|unmount]")
		os.Exit(1)
	}

	//create the volume provider
	var volp VolumeDriver
	volp = &DatasetVolumes{}

	//setup default output data
	output := Output{
		Status:  StatusNotSupported,
		Message: fmt.Sprintf("operation '%s' is unsupported", os.Args[1]),
	}

	//pass operations to the volume provider
	switch os.Args[1] {
	case OperationInit:
		output.Status = StatusSuccess
		output.Message = "Initialization successful"

		output.Capabilities, err = volp.Init()

	case OperationMount:
		output.Status = StatusSuccess
		output.Message = "Mount successful"

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
		output.Message = "Unmount successful"

		if len(os.Args) < 3 {
			err = fmt.Errorf("expected at least 3 arguments for unmount, got: %#v", os.Args)
		} else {
			err = volp.Unmount(os.Args[2])
		}
	}

	//if any operations returned an error, mark as failure
	if err != nil {
		log.Printf("failed to %+v: %v", os.Args, err)
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

//APIExt implements the DI interface
func (deps *Deps) APIExt() apiext.Interface {
	return nil
}
