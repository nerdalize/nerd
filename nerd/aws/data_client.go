package aws

import (
	"bytes"
	"crypto/sha256"
	"io"
	"io/ioutil"
	"path"

	"github.com/Sirupsen/logrus"
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

const (
	//uploadPolynomal is the polynomal that is used for chunked uploading.
	uploadPolynomal = 0x3DA3358B4DC173
	//NoOfRetries is the amount of retries when uploading or downloading to S3.
	NoOfRetries = 0
)

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
		return nil, errors.Wrap(err, "could not create AWS sessions")
	}
	return &DataClient{
		Session:          sess,
		DataClientConfig: conf,
	}, nil
}

//Upload uploads a piece of data.
func (client *DataClient) Upload(key string, body io.ReadSeeker) error {
	for i := 0; i <= NoOfRetries; i++ {
		svc := s3.New(client.Session)
		params := &s3.PutObjectInput{
			Bucket: aws.String(client.Bucket), // Required
			Key:    aws.String(key),           // Required
			Body:   body,
		}
		_, err := svc.PutObject(params)
		if err != nil {
			if i < NoOfRetries {
				continue
			}
			return errors.Wrapf(err, "could not put key %v", key)
		}
		break
	}
	return nil
}

//ChunkedUpload uploads data from a io.Reader (`r`) as a list of chunks. The Key of every chunk uploaded will be written to the KeyReadWriter (`kw`).
//ChunkedDownload reports its progress (the amount of bytes uploaded) to the progressCh.
//It will start a maximum of `concurrency` concurrent go routines to upload in paralllel.
//`root` is used as the root path of the chunk in S3. Root will be concatenated with the key to make the full S3 object path.
func (client *DataClient) ChunkedUpload(r io.Reader, kw data.KeyReadWriter, concurrency int, root string, progressCh chan<- int64) (err error) {
	cr := chunker.New(r, chunker.Pol(uploadPolynomal))
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
		exists, err := client.Exists(key) //check existence
		if err != nil {
			it.resCh <- &result{errors.Wrapf(err, "failed to check existence of '%x'", k), data.ZeroKey}
			return
		}

		if !exists {
			err = client.Upload(key, bytes.NewReader(it.chunk)) //if not exists put
			logrus.Debugf("Uploaded %s", key)
			if err != nil {
				it.resCh <- &result{errors.Wrapf(err, "failed to put chunk '%x'", k), data.ZeroKey}
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
			return errors.Wrapf(it.err, "failed to iterate")
		}

		res := <-it.resCh
		if res.err != nil {
			return res.err
		}

		err = kw.WriteKey(res.k)
		if err != nil {
			return errors.Wrapf(err, "failed to write key")
		}
	}

	return nil
}

//Download downloads a single object.
func (client *DataClient) Download(key string) (io.ReadCloser, error) {
	var r io.ReadCloser
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
			return nil, errors.Wrapf(err, "failed to download '%v'", key)
		}
		r = resp.Body
		break
	}
	return r, nil
}

//Exists checks if a given object key exists on S3.
func (client *DataClient) Exists(objectKey string) (has bool, err error) {
	svc := s3.New(client.Session)

	params := &s3.HeadObjectInput{
		Bucket: aws.String(client.Bucket), // Required
		Key:    aws.String(objectKey),
	}
	_, err = svc.HeadObject(params)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && (aerr.Code() == s3.ErrCodeNoSuchKey || aerr.Code() == sns.ErrCodeNotFoundException) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to check if key %v exists", objectKey)
	}
	return true, nil
}

//ChunkedDownload downloads a list of chunks and writes them to a io.Writer.
//ChunkedDownload reports its progress (the amount of bytes downloaded) to the progressCh.
//It will start a maximum of `concurrency` concurrent go routines to download in paralllel.
//`root` is used as the root path of the chunk in S3. Root will be concatenated with the Key read from `kr` to make the full S3 object path.
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
		r, err := client.Download(path.Join(root, it.k.ToString()))
		defer r.Close()
		if err != nil {
			it.resCh <- &result{errors.Wrapf(err, "failed to get key '%s'", it.k), nil}
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
			return errors.Wrapf(it.err, "failed to iterate")
		}

		res := <-it.resCh
		if res.err != nil {
			return res.err
		}

		_, err = cw.Write(res.chunk)
		if err != nil {
			return errors.Wrapf(err, "failed to write key")
		}
	}

	return nil
}
