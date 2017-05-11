package v1data

import (
	"context"
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

//Client holds a reference to an AWS session
type Client struct {
	DataOps DataOps
}

//DataOps is an interface to a set of data operations. The interface can be implemented to store / retrieve data from different data backends.
type DataOps interface {
	Upload(ctx context.Context, bucket, key string, body io.ReadSeeker) error
	Download(ctx context.Context, bucket, key string) (body io.ReadCloser, err error)
	Exists(ctx context.Context, bucket, key string) (exists bool, err error)
}

//NewClient creates a new data client that is capable of uploading and downloading (multiple) files.
func NewClient(ops DataOps) *Client {
	return &Client{ops}
}

//Upload uploads a single object.
func (c *Client) Upload(ctx context.Context, bucket, key string, body io.ReadSeeker) error {
	var err error
	for i := 0; i <= NoOfRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = c.DataOps.Upload(ctx, bucket, key, body)
			if err == nil {
				return nil
			}
		}
	}
	// TODO: Include logging.
	return client.NewError(fmt.Sprintf("failed to put '%v'", key), err)
}

//Download downloads a single object.
func (c *Client) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	var err error
	for i := 0; i <= NoOfRetries; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			var resp io.ReadCloser
			resp, err = c.DataOps.Download(ctx, bucket, key)
			if err == nil {
				return resp, nil
			}
		}
	}
	return nil, client.NewError(fmt.Sprintf("failed to download '%v'", key), err)
}

//Exists checks if a given object key exists.
func (c *Client) Exists(ctx context.Context, bucket, key string) (has bool, err error) {
	return c.DataOps.Exists(ctx, bucket, key)
}
