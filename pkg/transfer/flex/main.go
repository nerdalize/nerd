package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
type MountOptions struct {
	InputS3Key     string `json:"input/s3Key"`
	InputS3Bucket  string `json:"input/s3Bucket"`
	OutputS3Key    string `json:"output/s3Key"`
	OutputS3Bucket string `json:"output/s3Bucket"`
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
	Input  *transfer.Ref
	Output *transfer.Ref
}

func (volp *DatasetVolumes) writeDatasetOpts(mountPath string, opts MountOptions) (*datasetOpts, error) {
	dsopts := &datasetOpts{}
	if opts.InputS3Key != "" {
		dsopts.Input = &transfer.Ref{
			Key:    opts.InputS3Key,
			Bucket: opts.InputS3Bucket,
		}

		if dsopts.Input.Bucket == "" {
			return nil, errors.New("input key configured without a bucket")
		}
	}

	if opts.OutputS3Key != "" {
		dsopts.Output = &transfer.Ref{
			Key:    opts.OutputS3Key,
			Bucket: opts.OutputS3Bucket,
		}

		if dsopts.Output.Bucket == "" {
			return nil, errors.New("output key configured without a bucket")
		}
	}

	path := volp.getRelPath(mountPath, "json")
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

func (volp *DatasetVolumes) readDatasetOpts(mountPath string) (*datasetOpts, error) {
	path := volp.getRelPath(mountPath, "json")
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

func (volp *DatasetVolumes) deleteDatasetOpts(mountPath string) error {
	path := volp.getRelPath(mountPath, "json")
	err := os.Remove(path)
	return errors.Wrap(err, "failed to delete metadata file")
}

func (volp *DatasetVolumes) createFSInFile(mountPath string, size int64) (string, error) {
	volumePath := volp.getRelPath(mountPath, "volume")

	//Create file with room to contain writable file system
	f, err := os.Create(volumePath)
	if err != nil {
		err = errors.Wrap(err, "failed to create file system file")
		return volumePath, err
	}

	err = f.Truncate(size)
	if err != nil {
		err = errors.Wrap(err, "failed to allocate file system size")
		return volumePath, err
	}

	//Build file system within
	cmd := exec.Command("mkfs.ext4", volumePath)
	buf := bytes.NewBuffer(nil)
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		err = errors.Wrap(errors.New(strings.TrimSpace(buf.String())), "failed to execute mkfs command")
		return volumePath, err
	}

	return volumePath, nil
}

//@TODO: Parameters for this function need to be reconsidered, especially mountPath confusion
func (volp *DatasetVolumes) mountFSInFile(mountPath string, volumePath string) error {
	mountPoint := volp.getRelPath(mountPath, "mount")

	//Create mount point
	err := os.Mkdir(mountPoint, os.FileMode(0522))
	if err != nil {
		return errors.Wrap(err, "failed to create mount directory")
	}

	//Mount file system
	cmd := exec.Command("mount", volumePath, mountPoint)
	buf := bytes.NewBuffer(nil)
	cmd.Stderr = buf
	err = cmd.Run()
	if err != nil {
		return errors.Wrap(errors.New(strings.TrimSpace(buf.String())), "failed to execute mount command")
	}

	return nil
}

func (volp *DatasetVolumes) createOverlayFS(upperDir string, workDir string, lowerDir string, mountPoint string) error {
	overlayArgs := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerDir, upperDir, workDir)

	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", overlayArgs, mountPoint)
	buf := bytes.NewBuffer(nil)
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(errors.New(strings.TrimSpace(buf.String())), "failed to execute mount command")
	}

	return nil
}

func (volp *DatasetVolumes) getRelPath(mountPath string, name string) string {
	return filepath.Join(mountPath, "..", filepath.Base(mountPath)+"."+name)
}

//Deletes contents of a directory, but not the directory itself
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

//Init the flex volume
func (volp *DatasetVolumes) Init() (Capabilities, error) {
	return Capabilities{Attach: false}, nil
}

//Mount the flex voume, path: '/var/lib/kubelet/pods/c911e5f7-0392-11e8-8237-32f9813bbd5a/volumes/foo~cifs/input', opts: &main.MountOptions{FSType:"", PodName:"imagemagick", PodNamespace:"default", PodUID:"c911e5f7-0392-11e8-8237-32f9813bbd5a", PVOrVolumeName:"input", ReadWrite:"rw", ServiceAccountName:"default"}
func (volp *DatasetVolumes) Mount(mountPath string, opts MountOptions) (err error) {
	dsopts, err := volp.writeDatasetOpts(mountPath, opts)

	defer func() {
		if err != nil {
			volp.deleteDatasetOpts(mountPath)
		}
	}()

	if err != nil {
		return errors.Wrap(err, "failed to write volume database")
	}

	//Set up directory to hold base data
	basePath := volp.getRelPath(mountPath, "base")
	err = os.Mkdir(basePath, os.FileMode(0522))

	defer func() {
		if err != nil {
			os.RemoveAll(basePath)
		}
	}()

	if err != nil {
		return errors.Wrap(err, "failed to create input data directory")
	}

	// Download data to it if an input was specified
	if dsopts.Input != nil {
		var trans transfer.Transfer
		if trans, err = transfer.NewS3(&transfer.S3Conf{
			Bucket: dsopts.Input.Bucket,
		}); err != nil {
			return errors.Wrap(err, "failed to set up S3 transfer")
		}

		ref := &transfer.Ref{
			Bucket: dsopts.Input.Bucket,
			Key:    dsopts.Input.Key,
		}

		//@TODO when this fails flex volume retry mechanism will never succeed because the directory is not empty
		err = trans.Download(context.Background(), ref, basePath)
		if err != nil {
			return errors.Wrap(err, "failed to download data from S3")
		}
	}

	//Create volume to contain pod writes (hardcoded as 100 MB right now)
	volumePath, err := volp.createFSInFile(mountPath, 100*1024*1024)

	defer func() {
		if err != nil {
			os.RemoveAll(volumePath)
		}
	}()

	if err != nil {
		return errors.Wrap(err, "failed to create file system in a file")
	}

	//Mount the file system
	err = volp.mountFSInFile(mountPath, volumePath)

	defer func() {
		if err != nil {
			cmd := exec.Command("umount", volp.getRelPath(mountPath, "mount"))
			cmd.Run()

			os.RemoveAll(volp.getRelPath(mountPath, "mount"))
		}
	}()

	if err != nil {
		return errors.Wrap(err, "failed to mount file system in a file")
	}

	//Set up overlay file system using base and write restricted FS-in-file
	upperDir := filepath.Join(volp.getRelPath(mountPath, "mount"), "upper")
	workDir := filepath.Join(volp.getRelPath(mountPath, "mount"), "work")
	lowerDir := basePath

	err = os.Mkdir(upperDir, os.FileMode(0522))

	defer func() {
		if err != nil {
			os.RemoveAll(upperDir)
		}
	}()

	if err != nil {
		return errors.Wrap(err, "failed to make upper directory for overlayfs")
	}

	err = os.Mkdir(workDir, os.FileMode(0522))

	defer func() {
		if err != nil {
			os.RemoveAll(workDir)
		}
	}()

	if err != nil {
		return errors.Wrap(err, "failed to make work directory for overlayfs")
	}

	err = volp.createOverlayFS(upperDir, workDir, lowerDir, mountPath)

	defer func() {
		if err != nil {
			cmd := exec.Command("umount", mountPath)
			cmd.Run()

			volp.cleanDirectory(mountPath)
		}
	}()

	if err != nil {
		return errors.Wrap(err, "failed to mount overlayfs")
	}

	return nil
}

//Unmount the flex voume
func (volp *DatasetVolumes) Unmount(mountPath string) (err error) {
	var dsopts *datasetOpts
	dsopts, err = volp.readDatasetOpts(mountPath)
	if err != nil {
		return errors.Wrap(err, "failed to read volume database")
	}

	defer func() {
		//if there was no error during upload remove all data
		if err == nil {
			//Unmount overlay and FS-in-file
			cmd := exec.Command("umount", mountPath)
			err = cmd.Run()
			if err != nil {
				err = errors.Wrap(err, "failed to unmount overlayfs")
				return
			}

			cmd = exec.Command("umount", volp.getRelPath(mountPath, "mount"))
			err = cmd.Run()
			if err != nil {
				err = errors.Wrap(err, "failed to unmount file system in a file")
				return
			}

			err = os.Remove(volp.getRelPath(mountPath, "volume"))
			if err != nil {
				err = errors.Wrap(err, "failed to delete file system in a file")
				return
			}

			err = os.RemoveAll(volp.getRelPath(mountPath, "base"))
			if err != nil {
				err = errors.Wrap(err, "failed to delete input data")
				return
			}

			err = os.RemoveAll(volp.getRelPath(mountPath, "mount"))
			if err != nil {
				err = errors.Wrap(err, "failed to delete mount point for file system in a file")
				return
			}

			//Clean up everything else
			err = volp.cleanDirectory(mountPath)
			if err != nil {
				err = errors.Wrap(err, "failed to delete mount point")
				return
			}

			err = volp.deleteDatasetOpts(mountPath)
			if err != nil {
				err = errors.Wrap(err, "failed to delete dataset")
			}
		}
	}()

	if dsopts.Output == nil {
		return nil //no output dataset, do nothing with the volume data
	}

	var trans transfer.Transfer
	if trans, err = transfer.NewS3(&transfer.S3Conf{
		Bucket: dsopts.Output.Bucket,
	}); err != nil {
		err = errors.Wrap(err, "failed to set up S3 transfer")
		return err
	}

	ref := &transfer.Ref{
		Bucket: dsopts.Output.Bucket,
		Key:    dsopts.Output.Key,
	}

	_, err = trans.Upload(context.Background(), ref, mountPath)
	if err != nil {
		err = errors.Wrap(err, "failed to upload data to S3")
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
