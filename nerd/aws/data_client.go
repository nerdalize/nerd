package aws

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/nerdalize/nerd/nerd"
	"github.com/nerdalize/nerd/nerd/data"
	"github.com/pkg/errors"
	"github.com/restic/chunker"
)

//DirectoryPermissions are the permissions when a new directory is created upon file download.
const DirectoryPermissions = 0755

//KeyWriter writes a given key.
type KeyWriter interface {
	Write(k string) error
}

type KeyReader interface {
	Read() (k string, err error)
}

//DataClient holds a reference to an AWS session
type DataClient struct {
	Session *session.Session
	*DataClientConfig
}

//DataClientConfig provides config details to create a new DataClient.
type DataClientConfig struct {
	Credentials *credentials.Credentials
	Bucket      string
}

//NewDataClient creates a new data client that is capable of uploading and downloading (multiple) files.
func NewDataClient(conf *DataClientConfig) (*DataClient, error) {
	// TODO: Don't hardcode region
	sess, err := session.NewSession(&aws.Config{
		Credentials: conf.Credentials,
		Region:      aws.String(nerd.GetCurrentUser().Region),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create AWS sessions: %v", err)
	}
	return &DataClient{
		Session:          sess,
		DataClientConfig: conf,
	}, nil
}

//Upload uploads a piece of data.
func (client *DataClient) Upload(key string, body io.ReadSeeker) error {
	// TODO: retries
	svc := s3.New(client.Session)
	params := &s3.PutObjectInput{
		Bucket: aws.String(client.Bucket), // Required
		Key:    aws.String(key),           // Required
		Body:   body,
	}
	_, err := svc.PutObject(params)
	if err != nil {
		return errors.Wrapf(err, "could not put key %v", key)
	}
	return nil
}

func (client *DataClient) ChunkedUpload(r io.Reader, kw data.KeyReadWriter, concurrency int, root string, progressCh chan<- int64) (err error) {
	cr := chunker.New(r, chunker.Pol(0x3DA3358B4DC173))
	type result struct {
		err error
		k   data.Key
	}

	type item struct {
		chunk []byte
		size  int64
		resCh chan *result
		err   error
	}

	work := func(it *item) {
		k := data.Key(sha256.Sum256(it.chunk)) //hash
		key := path.Join(root, k.ToString())
		exists, err := client.Has(key) //check existence
		if err != nil {
			it.resCh <- &result{fmt.Errorf("failed to check existence of '%x': %v", k, err), data.ZeroKey}
			return
		}

		if !exists {
			err = client.Upload(key, bytes.NewReader(it.chunk)) //if not exists put
			if err != nil {
				it.resCh <- &result{fmt.Errorf("failed to put chunk '%x': %v", k, err), data.ZeroKey}
				return
			}
		}
		progressCh <- int64(len(it.chunk))

		it.resCh <- &result{nil, k}
	}

	//fan out
	itemCh := make(chan *item, concurrency)
	go func() {
		defer close(itemCh)
		buf := make([]byte, chunker.MaxSize)
		for {
			chunk, err := cr.Next(buf)
			if err != nil {
				if err != io.EOF {
					itemCh <- &item{err: err}
				}
				break
			}

			it := &item{
				chunk: make([]byte, chunk.Length),
				resCh: make(chan *result),
			}

			copy(it.chunk, chunk.Data) //underlying buffer is switched out

			go work(it)  //create work
			itemCh <- it //send to fan-in thread for syncing results
		}
	}()

	//fan-in
	for it := range itemCh {
		if it.err != nil {
			return fmt.Errorf("failed to iterate: %v", it.err)
		}

		res := <-it.resCh
		if res.err != nil {
			return res.err
		}

		err = kw.WriteKey(res.k)
		if err != nil {
			return fmt.Errorf("failed to write key: %v", err)
		}
	}

	return nil
}

//Download downloads a single file.
func (client *DataClient) Download(key string) (io.ReadCloser, error) {
	var r io.ReadCloser
	NoOfRetries := 2
	for i := 0; i <= NoOfRetries; i++ {
		svc := s3.New(client.Session)
		params := &s3.GetObjectInput{
			Bucket: aws.String(client.Bucket), // Required
			Key:    aws.String(key),           // Required
		}
		resp, err := svc.GetObject(params)

		if err != nil {
			if i < NoOfRetries {
				continue
			}
			// TODO: fmt should be errors
			return nil, fmt.Errorf("failed to download '%v': %v", key, err)
		}
		r = resp.Body
		break
	}
	return r, nil
}

func (client *DataClient) Has(key string) (has bool, err error) {
	svc := s3.New(client.Session)

	params := &s3.HeadObjectInput{
		Bucket: aws.String(client.Bucket), // Required
		Key:    aws.String(key),
	}
	_, err = svc.HeadObject(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && (aerr.Code() == s3.ErrCodeNoSuchKey || aerr.Code() == sns.ErrCodeNotFoundException) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to check if key %v exists", key)
	}
	return true, nil
}

func (client *DataClient) ChunkedDownload(kr data.KeyReadWriter, cw io.Writer, concurrency int, root string, progressCh chan<- int64) (err error) {
	type result struct {
		err   error
		chunk []byte
	}

	type item struct {
		k     data.Key
		resCh chan *result
		err   error
	}

	work := func(it *item) {
		// TODO: add root
		r, err := client.Download(path.Join(root, it.k.ToString()))
		defer r.Close()
		if err != nil {
			it.resCh <- &result{fmt.Errorf("failed to get key '%s': %v", it.k, err), nil}
			return
		}

		chunk, err := ioutil.ReadAll(r)
		if err != nil {
			it.resCh <- &result{errors.Wrap(err, "failed to copy chunk to byte buffer"), nil}
		}

		progressCh <- int64(len(chunk))

		it.resCh <- &result{nil, chunk}
	}

	//fan out
	itemCh := make(chan *item, concurrency)
	go func() {
		defer close(itemCh)
		for {
			k, err := kr.ReadKey()
			if err != nil {
				if err != io.EOF {
					itemCh <- &item{err: err}
				}
				break
			}

			it := &item{
				k:     k,
				resCh: make(chan *result),
			}

			go work(it)  //create work
			itemCh <- it //send to fan-in thread for syncing results
		}
	}()

	//fan-in
	for it := range itemCh {
		if it.err != nil {
			return fmt.Errorf("failed to iterate: %v", it.err)
		}

		res := <-it.resCh
		if res.err != nil {
			return res.err
		}

		_, err = cw.Write(res.chunk)
		if err != nil {
			return fmt.Errorf("failed to write key: %v", err)
		}
	}

	return nil
}
