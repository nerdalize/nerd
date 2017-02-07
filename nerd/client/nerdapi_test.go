// +build integration

package client

import (
	"fmt"
	"os"
	"testing"

	flags "github.com/jessevdk/go-flags"
	"github.com/nerdalize/nerd/command"
	"github.com/nerdalize/nerd/nerd/client"
)

func newClient() (*NerdAPIClient, error) {
	opts := &command.NerdAPIOpts{}
	args, err := flags.ParseArgs(opts, os.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse flags: %v", err)
	}
	c := client.NewNerdAPI(client.NerdAPIConfig{
		Scheme:   cmd.opts.NerdAPIScheme,
		Host:     cmd.opts.NerdAPIHostname,
		BasePath: cmd.opts.NerdAPIBasePath,
		Version:  cmd.opts.NerdAPIVersion,
	})
	return c, nil
}

func TestRun(t *testing.T) {
	c, err := newClient()
	if err != nil {
		t.Error(err)
		return
	}

	image := "busybox"
	dataset := "test-dataset"
	awsAccessKey := "12345"
	awsSecret := "67890"
	args := []string{}
	err = c.Run(image, dataset, awsAccessKey, awsSecret, args)
	if err != nil {
		t.Errorf("failed to create task %v: %v", image, err)
	}
}
