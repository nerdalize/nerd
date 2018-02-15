package transfer

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
	upl  s3manageriface.UploaderAPI
	dwn  s3manageriface.DownloaderAPI
	api  s3iface.S3API
}

//S3StoreConfig allows configuration of the S3 object store implementation
type S3StoreConfig struct {
	Prefix       string
	Region       string
	AccessKey    string
	SecretKey    string
	SessionToken string
}

//CreateS3Store is the factory method for the s3 store
func CreateS3Store(opts map[string]string) (store Store, err error) {
	bucket, ok := opts["aws_s3_bucket"]
	if !ok {
		return nil, errors.New("aws_s3_bucket configuration is missing")
	}

	cfg := S3StoreConfig{}
	cfg.Prefix, _ = opts["aws_s3_prefix"]
	cfg.Region, _ = opts["aws_s3_region"]
	cfg.AccessKey, _ = opts["aws_s3_access_key"]
	cfg.SecretKey, _ = opts["aws_s3_secret_key"]
	cfg.SessionToken, _ = opts["aws_s3_session_token"]

	return NewS3Store(bucket, cfg)
}

//NewS3Store creates an s3 implementation of the object store
func NewS3Store(bucket string, cfg S3StoreConfig) (store *S3Store, err error) {
	store = &S3Store{
		bucket: bucket,
		prefix: cfg.Prefix,
	}

	if store.prefix != "" && !strings.HasSuffix(store.prefix, "/") {
		return nil, errors.Errorf("store prefix must end with a forward slash")
	}

	if cfg.Region == "" {
		cfg.Region = endpoints.UsEast1RegionID //this will make the sdk use the global s3 endpoint
	}

	awscfg := &aws.Config{Region: aws.String(cfg.Region)}
	if cfg.AccessKey != "" { //if we have some credentials, configure the session as such
		awscfg.Credentials = credentials.NewStaticCredentials(
			cfg.AccessKey, cfg.SecretKey, cfg.SessionToken,
		)
	}

	var sess *session.Session
	if sess, err = session.NewSession(awscfg); err != nil {
		return nil, errors.Wrapf(err, "failed to create AWS session")
	}

	s3api := s3.New(sess)
	if cfg.AccessKey == "" { //without credentials we'll disable request signing
		s3api.Handlers.Sign.Clear()
		s3api.Handlers.Sign.PushBackNamed(corehandlers.BuildContentLengthHandler)
		//we delibrately don't add actual signing middleware for anonymous access
	}

	store.upl = s3manager.NewUploaderWithClient(s3api)
	store.dwn = s3manager.NewDownloaderWithClient(s3api)
	store.api = s3api

	return store, nil
}

//Get a object from the store with key 'k'
func (store *S3Store) Get(ctx context.Context, k string, w io.WriterAt) (err error) {
	if _, err = store.dwn.DownloadWithContext(ctx, w, &s3.GetObjectInput{
		Bucket: aws.String(store.bucket),
		Key:    aws.String(k),
	}); err != nil {
		return errors.Wrapf(err, "failed to download object")
	}

	return nil
}

//Put an object into the store at key 'k'
func (store *S3Store) Put(ctx context.Context, k string, r io.Reader) (err error) {
	if _, err := store.upl.UploadWithContext(ctx, &s3manager.UploadInput{
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
