package v1data

import (
	"fmt"
	"io"

	"github.com/nerdalize/nerd/nerd/client"
)

const (
	//LogGroup is the group for each log statement.
	LogGroup = "DataClient"
	//NoOfRetries is the amount of retries when uploading or downloading to S3.
	NoOfRetries = 2
)

//DataClient holds a reference to an AWS session
type Client struct {
	DataOps DataOps
}

//DataOps is an interface to a set of data operations. The interface can be implemented to store / retrieve data from different data backends.
type DataOps interface {
	Upload(bucket, key string, body io.ReadSeeker) error
	Download(bucket, key string) (body io.ReadCloser, err error)
	Exists(bucket, key string) (exists bool, err error)
}

//NewClient creates a new data client that is capable of uploading and downloading (multiple) files.
func NewClient(ops DataOps) *Client {
	return &Client{ops}
}

//Upload uploads a single object.
func (c *Client) Upload(bucket, key string, body io.ReadSeeker) error {
	for i := 0; i <= NoOfRetries; i++ {
		err := c.DataOps.Upload(bucket, key, body)
		if err != nil {
			if i < NoOfRetries {
				continue
			}
			return client.NewError(fmt.Sprintf("failed to put '%v'", key), err)
		}
		break
	}
	// TODO: Include logging.
	return nil
}

//Download downloads a single object.
func (c *Client) Download(bucket, key string) (io.ReadCloser, error) {
	var r io.ReadCloser
	for i := 0; i <= NoOfRetries; i++ {
		resp, err := c.DataOps.Download(bucket, key)
		if err != nil {
			if i < NoOfRetries {
				continue
			}
			return nil, client.NewError(fmt.Sprintf("failed to download '%v'", key), err)
		}
		r = resp
		break
	}
	return r, nil
}

//Exists checks if a given object key exists.
func (c *Client) Exists(bucket, key string) (has bool, err error) {
	return c.DataOps.Exists(bucket, key)
}
