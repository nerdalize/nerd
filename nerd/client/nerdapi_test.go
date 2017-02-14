// +build integration

package client_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/nerdalize/nerd/nerd/client"
)

func newClient() (*client.NerdAPIClient, error) {
	fullUrl := os.Getenv("TEST_NERD_API_FULL_URL")
	version := os.Getenv("TEST_NERD_API_VERSION")
	c, err := client.NewNerdAPIFromURL(fullUrl, version)
	if err != nil {
		return nil, fmt.Errorf("failed to create nerd API client: %v", err)
	}
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
	err = c.CreateTask(image, dataset, awsAccessKey, awsSecret, args)
	if err != nil {
		t.Errorf("failed to create task %v: %v", image, err)
	}
}
