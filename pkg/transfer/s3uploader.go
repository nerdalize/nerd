package transfer

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/corehandlers"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
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
	if cfg.Bucket == "" {
		return nil, errors.New("no bucket configured")
	}

	upl = &S3Uploader{cfg: cfg}
	if cfg.Region == "" {
		cfg.Region = endpoints.UsEast1RegionID //this will make the sdk use the global s3 endpoint
	}

	awscfg := &aws.Config{Region: aws.String(cfg.Region)}
	if cfg.AccessKey != "" { //if we have some credentials, configure the session as such
		awscfg.Credentials = credentials.NewStaticCredentials(
			cfg.AccessKey, cfg.SecretKey, cfg.SessionToken,
		)
	}

	if upl.sess, err = session.NewSession(awscfg); err != nil {
		return nil, err
	}

	s3api := s3.New(upl.sess)
	if cfg.AccessKey == "" { //without credentials we'll disable request signing
		s3api.Handlers.Sign.Clear()
		s3api.Handlers.Sign.PushBackNamed(corehandlers.BuildContentLengthHandler)
		//we delibrately don't add actual signing middleware
	}

	//setup the official uploader
	upl.upl = s3manager.NewUploaderWithClient(s3api)
	return upl, nil
}

//Upload data at a local path to the remote storage and return a reference
func (upl *S3Uploader) Upload(path string) (r *Ref, err error) {
	buf := bytes.NewBuffer(nil)
	zipw := zip.NewWriter(buf)
	if err = func() error {
		defer zipw.Close()
		return filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
			if p == path || fi.IsDir() {
				return nil //skip dirs
			}

			rel, err := filepath.Rel(path, p)
			if err != nil {
				return errors.Wrap(err, "failed to determine relative path")
			}

			f, err := os.Open(p)
			if err != nil {
				return errors.Wrap(err, "failed to open file")
			}

			defer f.Close()
			zipf, err := zipw.Create(rel)
			if err != nil {
				return errors.Wrap(err, "failed to create zip file")
			}

			_, err = io.Copy(zipf, f)
			if err != nil {
				return errors.Wrap(err, "failed to copy file")
			}

			return nil
		})
	}(); err != nil {
		return nil, errors.Wrap(err, "failed to create zip file")
	}

	out, err := upl.upl.Upload(&s3manager.UploadInput{
		Bucket: aws.String(upl.cfg.Bucket),
		Key:    aws.String(uuid.NewV4().String() + ".zip"),
		Body:   buf,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to perform upload")
	}

	return &Ref{Location: out.Location}, nil
}
