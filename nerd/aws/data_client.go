package aws

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/pkg/errors"
)

//DataClient is a client to AWS' S3 service. The client implements the v1data.DataOps interface.
type DataClient struct {
	Service *s3.S3
}

//NewDataClient creates a new DataClient.
func NewDataClient(c *credentials.Credentials, region string) (*DataClient, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: c,
		Region:      aws.String(region),
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not create AWS sessions")
	}
	return &DataClient{
		Service: s3.New(sess),
	}, nil
}

//Upload uploads an object to S3.
func (c *DataClient) Upload(bucket, key string, body io.ReadSeeker) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket), // Required
		Key:    aws.String(key),    // Required
		Body:   body,
	}
	_, err := c.Service.PutObject(input)
	return err
}

//Download downloads an object to S3.
func (c *DataClient) Download(bucket, key string) (body io.ReadCloser, err error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket), // Required
		Key:    aws.String(key),    // Required
	}
	resp, err := c.Service.GetObject(input)

	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

//Exists checks whether an object exists on S3.
func (c *DataClient) Exists(bucket, key string) (exists bool, err error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket), // Required
		Key:    aws.String(key),
	}
	_, err = c.Service.HeadObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && (aerr.Code() == s3.ErrCodeNoSuchKey || aerr.Code() == sns.ErrCodeNotFoundException) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to check if key %v exists", key)
	}
	return true, nil
}
