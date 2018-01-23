package transfer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

//S3Uploader encapsulates logic for uploading a local directory
//to S3 object storage
type S3Uploader struct {
	sess *session.Session
	cfg  *S3Conf
	upl  *s3manager.Uploader
}

//NewS3Uploader creates an S3 uploader
func NewS3Uploader(cfg *S3Conf) (upl *S3Uploader, err error) {
	upl = &S3Uploader{cfg: cfg}
	if upl.sess, err = session.NewSession(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewSharedCredentials("", "test-account"),
	}); err != nil {
		return nil, err
	}

	return upl, nil
}
