package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

//MountOptions is specified whenever Kubernetes calls the mount
type MountOptions struct {
	FSType             string `json:"kubernetes.io/fsType"`
	PodName            string `json:"kubernetes.io/pod.name"`
	PodNamespace       string `json:"kubernetes.io/pod.namespace"`
	PodUID             string `json:"kubernetes.io/pod.uid"`
	PVOrVolumeName     string `json:"kubernetes.io/pvOrVolumeName"`
	ReadWrite          string `json:"kubernetes.io/readwrite"`
	ServiceAccountName string `json:"kubernetes.io/serviceAccount.name"`
}

//Capabilities of the flex volume
type Capabilities struct {
	Attach bool `json:"attach"`
}

func runInit() (Capabilities, error) {
	return Capabilities{Attach: false}, nil
}

func runMount(path string, opts *MountOptions) error {
	return fmt.Errorf("path: '%v', opts: %#v", path, opts)
}

func runUnmount(path string) error {
	return fmt.Errorf("path: '%v'", path)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: nerd-flex-volume [init|mount|unmount]")
	}

	err := fmt.Errorf("operation '%s' is unsupported", os.Args[1])
	output := Output{}
	switch os.Args[1] {
	case "init":
		output.Capabilities, err = runInit()

	case "mount":
		if len(os.Args) < 4 {
			err = fmt.Errorf("expected at least 4 arguments for mount, got: %#v", os.Args)
		} else {
			opts := &MountOptions{}
			err = json.Unmarshal([]byte(os.Args[3]), opts)
			if err == nil {
				err = runMount(os.Args[2], opts)
			}
		}

	case "unmount":
		if len(os.Args) < 3 {
			err = fmt.Errorf("expected at least 3 arguments for unmount, got: %#v", os.Args)
		} else {
			err = runUnmount(os.Args[2])
		}
	}

	if err != nil {
		output.Status = StatusFailure
		output.Message = err.Error()
	}

	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(output)
	if err != nil {
		log.Fatalf("failed to encode output: %v", err)
	}
}
