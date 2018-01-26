package transfer

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"io/ioutil"
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

//S3 encapsulates logic for uploading a local directory
//to S3 object storage
type S3 struct {
	sess *session.Session
	cfg  *S3Conf
	upl  *s3manager.Uploader
	dwn  *s3manager.Downloader
}

//NewS3 creates an S3 transfer
func NewS3(cfg *S3Conf) (trans *S3, err error) {
	if cfg.Bucket == "" {
		return nil, errors.New("no bucket configured")
	}

	trans = &S3{cfg: cfg}
	if cfg.Region == "" {
		cfg.Region = endpoints.UsEast1RegionID //this will make the sdk use the global s3 endpoint
	}

	awscfg := &aws.Config{Region: aws.String(cfg.Region)}
	if cfg.AccessKey != "" { //if we have some credentials, configure the session as such
		awscfg.Credentials = credentials.NewStaticCredentials(
			cfg.AccessKey, cfg.SecretKey, cfg.SessionToken,
		)
	}

	if trans.sess, err = session.NewSession(awscfg); err != nil {
		return nil, err
	}

	s3api := s3.New(trans.sess)
	if cfg.AccessKey == "" { //without credentials we'll disable request signing
		s3api.Handlers.Sign.Clear()
		s3api.Handlers.Sign.PushBackNamed(corehandlers.BuildContentLengthHandler)
		//we delibrately don't add actual signing middleware
	}

	//setup the official uploader
	trans.upl = s3manager.NewUploaderWithClient(s3api)
	trans.dwn = s3manager.NewDownloaderWithClient(s3api)
	return trans, nil
}

//Download a data reference to the path provided
func (trans *S3) Download(ctx context.Context, r *Ref, path string) (err error) {
	dir, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to open directory")
		}

		err = os.Mkdir(path, 0777) //@TODO decide on permissions before umask
		if err != nil {
			return errors.Wrap(err, "failed to create directory")
		}

		dir, err = os.Open(path)
		if err != nil {
			return errors.Wrap(err, "failed open created directory")
		}
	}

	fis, err := dir.Readdirnames(1)
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "failed to read directory")
	}

	if len(fis) > 0 {
		return errors.New("directory is not empty")
	}

	f, err := ioutil.TempFile("", r.Key)
	if err != nil {
		return errors.Wrap(err, "failed to create temp file")
	}

	defer f.Close()
	defer os.Remove(f.Name())

	var size int64
	if size, err = trans.dwn.DownloadWithContext(ctx, f, &s3.GetObjectInput{
		Bucket: aws.String(r.Bucket),
		Key:    aws.String(r.Key),
	}); err != nil {
		return errors.Wrap(err, "failed to download object")
	}

	zipr, err := zip.NewReader(f, size)
	if err != nil {
		return errors.Wrap(err, "failed to open zip reader")
	}

	for _, zipf := range zipr.File {
		if err = func() error {
			var rc io.ReadCloser
			rc, err = zipf.Open()
			if err != nil {
				return errors.Wrap(err, "failed to open zip file")
			}

			defer rc.Close()
			var f *os.File
			f, err = os.Create(filepath.Join(path, zipf.Name)) //@TODO what permissions(?) executable bits?
			if err != nil {
				return errors.Wrap(err, "failed to create file to extract to")
			}

			defer f.Close()
			_, err = io.Copy(f, rc)
			if err != nil {
				return errors.Wrap(err, "failed to copy zip file contents")
			}

			return nil
		}(); err != nil {
			return err
		}
	}

	return nil
}

//Upload data at a local path to the remote storage and return a reference
func (trans *S3) Upload(ctx context.Context, path string) (size int, r *Ref, err error) {
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
		return 0, nil, errors.Wrap(err, "failed to create zip file")
	}

	uid, err := uuid.NewV4()
	if err != nil {
		return 0, nil, errors.Wrap(err, "failed to create uuid")
	}

	size = buf.Len()
	key := uid.String() + ".zip"
	if _, err = trans.upl.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(trans.cfg.Bucket),
		Key:    aws.String(key),
		Body:   buf,
	}); err != nil {
		return 0, nil, errors.Wrap(err, "failed to perform upload")
	}

	return size, &Ref{
		Bucket: trans.cfg.Bucket,
		Key:    key,
	}, nil
}
