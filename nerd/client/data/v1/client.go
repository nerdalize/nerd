package v1data

import (
	"io"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/pkg/errors"
)

const (
	//LogGroup is the group for each log statement.
	LogGroup = "DataClient"
	//uploadPolynomal is the polynomal that is used for chunked uploading.
	uploadPolynomal = 0x3DA3358B4DC173
	//NoOfRetries is the amount of retries when uploading or downloading to S3.
	NoOfRetries = 2
)

//DataClient holds a reference to an AWS session
type Client struct {
	DataOps DataOps
}

type DataOps interface {
	PutObject(bucket, key string, body io.ReadSeeker) error
	GetObject(bucket, key string) (body io.ReadCloser, err error)
	HeadObject(bucket, key string) error
}

//NewDataClient creates a new data client that is capable of uploading and downloading (multiple) files.
func NewDataClient(ops DataOps) *Client {
	return &Client{ops}
}

//Upload uploads a piece of data.
func (c *Client) Upload(bucket, key string, body io.ReadSeeker) error {
	for i := 0; i <= NoOfRetries; i++ {
		err := c.DataOps.PutObject(bucket, key, body)
		if err != nil {
			if i < NoOfRetries {
				continue
			}
			return errors.Wrapf(err, "could not put key %v", key)
		}
		break
	}
	// logrus.WithFields(logrus.Fields{
	// 	"group":  LogGroup,
	// 	"action": "upload",
	// 	"key":    key,
	// }).Debugf("Uploaded %s", key)
	return nil
}

//Download downloads a single object.
func (c *Client) Download(bucket, key string) (io.ReadCloser, error) {
	var r io.ReadCloser
	for i := 0; i <= NoOfRetries; i++ {
		resp, err := c.DataOps.GetObject(bucket, key)
		if err != nil {
			if i < NoOfRetries {
				continue
			}
			return nil, errors.Wrapf(err, "failed to download '%v'", key)
		}
		r = resp
		break
	}
	// logrus.WithFields(logrus.Fields{
	// 	"group":  LogGroup,
	// 	"action": "download",
	// 	"key":    key,
	// }).Debugf("Downloaded %s", key)
	return r, nil
}

//Exists checks if a given object key exists on S3.
func (c *Client) Exists(bucket, key string) (has bool, err error) {
	err = c.DataOps.HeadObject(bucket, key)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && (aerr.Code() == s3.ErrCodeNoSuchKey || aerr.Code() == sns.ErrCodeNotFoundException) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to check if key %v exists", key)
	}
	return true, nil
}
