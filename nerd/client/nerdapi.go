package client

import (
	"fmt"
	"net/url"
	"path"

	"github.com/dghubble/sling"
	"github.com/nerdalize/nerd/nerd/payload"
)

const (
	defaultScheme   = "https"
	defaultHost     = "platform.nerdalize.net"
	defaultBasePath = ""
	defaultVersion  = "v1"
)

type NerdAPIClient struct {
	// client *REST
	NerdAPIConfig
	sl *sling.Sling
}

type NerdAPIConfig struct {
	Scheme   string
	Host     string
	BasePath string
	Version  string
}

func NewNerdAPI(config NerdAPIConfig) *NerdAPIClient {
	if config.Scheme == "" {
		config.Scheme = defaultScheme
	}
	if config.Host == "" {
		config.Host = defaultHost
	}
	if config.BasePath == "" {
		config.BasePath = defaultBasePath
	}
	if config.Version == "" {
		config.Version = defaultVersion
	}
	return &NerdAPIClient{
		NerdAPIConfig: config,
	}
}

func (nerdapi *NerdAPIClient) url(p string) string {
	url := &url.URL{
		Scheme: nerdapi.Scheme,
		Host:   nerdapi.Host,
		Path:   path.Join(nerdapi.BasePath, p),
		//TODO: include version
		// Path:   path.Join(nerdapi.BasePath, nerdapi.Version, p),
	}
	return url.String()
}

func (nerdapi *NerdAPIClient) Run(image string, dataset string, awsAccessKey string, awsSecret string, args []string) error {
	// set env variables
	args = append(args, "-e=DATASET="+dataset)
	args = append(args, "-e=AWS_ACCESS_KEY_ID="+awsAccessKey)
	args = append(args, "-e=AWS_SECRET_ACCESS_KEY="+awsSecret)

	// create payload
	p := &payload.Task{
		Image:   image,
		Dataset: dataset,
		Args:    args,
	}

	// post request
	url := nerdapi.url("/tasks")
	resp, err := sling.New().
		Post(url).
		BodyJSON(p).
		ReceiveSuccess(nil)

	if err != nil {
		return fmt.Errorf("failed to send request to %v (POST): %v", url, err)
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API request '%s (POST)' returned unexpected status from API: %v", url, resp.Status)
	}

	return nil
}

func (nerdapi *NerdAPIClient) Status() (s []payload.Task, err error) {
	url := nerdapi.url("/tasks")
	resp, err := sling.New().
		Get(url).
		ReceiveSuccess(&s)

	if err != nil {
		return []payload.Task{}, fmt.Errorf("failed to send request to %v (GET): %v", url, err)
	}
	if resp.StatusCode >= 400 {
		return []payload.Task{}, fmt.Errorf("API request '%s (GET)' returned unexpected status from API: %v", url, resp.Status)
	}
	return
}
