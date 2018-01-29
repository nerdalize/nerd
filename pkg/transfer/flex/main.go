package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nerdalize/nerd/pkg/transfer"
)

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
type MountOptions map[string]string

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

//DatasetType determines if the volume will be uploaded or downloaded
type DatasetType string

const (
	//DatasetTypeInput will be downloaded
	DatasetTypeInput = "input"

	//DatasetTypeOutput will be uploaded
	DatasetTypeOutput = "output"
)

type datasetOpts struct {
	Type   DatasetType
	Key    string
	Bucket string
}

func (volp *DatasetVolumes) writeDatasetOpts(mountPath string, opts MountOptions) (*datasetOpts, error) {
	dsopts := &datasetOpts{}
	typ, _ := opts["type"]
	switch typ {
	case DatasetTypeInput:
		dsopts.Type = DatasetTypeInput
	case DatasetTypeOutput:
		dsopts.Type = DatasetTypeOutput
	default:
		return nil, fmt.Errorf("unsupported dataset type specified: '%s'", typ)
	}

	dsopts.Key, _ = opts["key"]
	if dsopts.Key == "" {
		return nil, fmt.Errorf("no dataset key configured for volume")
	}

	dsopts.Bucket, _ = opts["bucket"]
	if dsopts.Bucket == "" {
		return nil, fmt.Errorf("no bucket key configured for volume")
	}

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

	if dsopts.Type != DatasetTypeInput {
		return nil //not an input dataset, do nothing on mount
	}

	var trans transfer.Transfer
	if trans, err = transfer.NewS3(&transfer.S3Conf{
		Bucket: dsopts.Bucket,
	}); err != nil {
		return err
	}

	ref := &transfer.Ref{
		Bucket: dsopts.Bucket,
		Key:    dsopts.Key,
	}

	err = trans.Download(context.Background(), ref, mountPath)
	if err != nil {
		return err
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

	if dsopts.Type != DatasetTypeOutput {
		return nil //not an output dataset, do nothing on unmount
	}

	var trans transfer.Transfer
	if trans, err = transfer.NewS3(&transfer.S3Conf{
		Bucket: dsopts.Bucket,
	}); err != nil {
		return err
	}

	ref := &transfer.Ref{
		Bucket: dsopts.Bucket,
		Key:    dsopts.Key,
	}

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
