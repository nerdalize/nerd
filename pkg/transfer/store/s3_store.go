package transferstore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/corehandlers"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
	"github.com/pkg/errors"
)

//S3Store provides an S3 Backed store
type S3Store struct {
	bucket string
	prefix string

	sess *session.Session
	dwn  s3manageriface.DownloaderAPI
	api  s3iface.S3API
}

//NewS3Store creates an s3 implementation of the object store
func NewS3Store(cfg StoreOptions) (store *S3Store, err error) {
	store = &S3Store{
		bucket: cfg.S3StoreBucket,
		prefix: cfg.S3StorePrefix,
	}

	if store.prefix != "" && !strings.HasSuffix(store.prefix, "/") {
		return nil, errors.Errorf("store prefix must end with a forward slash")
	}

	if cfg.S3StoreAWSRegion == "" {
		cfg.S3StoreAWSRegion = endpoints.UsEast1RegionID //this will make the sdk use the global s3 endpoint
	}

	awscfg := &aws.Config{Region: aws.String(cfg.S3StoreAWSRegion)}
	if cfg.S3StoreAWSRegion != "" { //if we have some credentials, configure the session as such
		awscfg.Credentials = credentials.NewStaticCredentials(
			cfg.S3StoreAccessKey, cfg.S3StoreSecretKey, cfg.S3SessionToken,
		)
	}

	var sess *session.Session
	if sess, err = session.NewSession(awscfg); err != nil {
		return nil, errors.Wrapf(err, "failed to create AWS session")
	}

	s3api := s3.New(sess)
	if cfg.S3StoreAccessKey == "" { //without credentials we'll disable request signing
		s3api.Handlers.Sign.Clear()
		s3api.Handlers.Sign.PushBackNamed(corehandlers.BuildContentLengthHandler)
		//we delibrately don't add actual signing middleware for anonymous access
	}

	store.dwn = s3manager.NewDownloaderWithClient(s3api)
	store.api = s3api

	return store, nil
}

//Head returns metadata for the object
func (store *S3Store) Head(ctx context.Context, k string) (size int64, err error) {
	var out *s3.HeadObjectOutput
	if out, err = store.api.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(store.bucket),
		Key:    aws.String(k),
	}); err != nil {
		return 0, errors.Wrapf(err, "failed to download object")
	}

	size = aws.Int64Value(out.ContentLength)
	return size, nil
}

//Get a object from the store with key 'k' and write it to 'w'
func (store *S3Store) Get(ctx context.Context, k string, w io.WriterAt) (err error) {
	if _, err = store.dwn.DownloadWithContext(ctx, w, &s3.GetObjectInput{
		Bucket: aws.String(store.bucket),
		Key:    aws.String(k),
	}); err != nil {
		return errors.Wrapf(err, "failed to download object")
	}

	return nil
}

//Put an object into the store at key 'k' by reading from 'r'
func (store *S3Store) Put(ctx context.Context, k string, r io.ReadSeeker) (err error) {
	if _, err := store.api.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Body:   r,
		Bucket: aws.String(store.bucket),
		Key:    aws.String(k),
	}); err != nil {
		return errors.Wrap(err, "failed to download object")
	}

	return nil
}

//Del will remove an object from the store at key 'k'
func (store *S3Store) Del(ctx context.Context, k string) error {
	if _, err := store.api.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(store.bucket),
		Key:    aws.String(k),
	}); err != nil {
		return errors.Wrap(err, "failed to delete object")
	}

	return nil
}

//TempS3Bucket creates a temporary s3 bucket that can be removed again
//by calling clean(). This is mainly usefull for testing purposes throughout
//the codebase of this project. The name will be a randomly generated name
//prefixed with 'nerd-tests-'
func TempS3Bucket() (name string, clean func(), err error) {
	s3api := s3.New(session.Must(session.NewSession()))

	d := make([]byte, 16)
	_, err = rand.Read(d)
	if err != nil {
		return "", nil, errors.Wrap(err, "failed to read random bytes")
	}

	name = fmt.Sprintf("nerd-tests-%s", hex.EncodeToString(d))
	if _, err = s3api.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(name),
	}); err != nil {
		return "", nil, errors.Wrap(err, "failed to create bucket")
	}

	return name, func() {
		_, err = s3api.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(name),
		})

		if err != nil {
			panic(err)
		}
	}, nil
}
