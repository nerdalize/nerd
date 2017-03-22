package aws

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/nerdalize/nerd/nerd"
	"github.com/pkg/errors"
	"github.com/restic/chunker"
)

//DirectoryPermissions are the permissions when a new directory is created upon file download.
const DirectoryPermissions = 0755

//KeyWriter writes a given key.
type KeyWriter interface {
	Write(k string) error
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

//UploadFile uploads a single file.
func (client *DataClient) UploadFile(filePath string, key string, root string) error {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("could not open file '%v': %v", filePath, err)
	}
	err = client.Upload(path.Join(root, key), file)
	if err != nil {
		return errors.Wrapf(err, "failed to upload file %v", filePath)
	}
	return nil
}

//UploadDir uploads every single file in the directory and all its subdirectories.
func (client *DataClient) UploadDir(dir string, root string, kw KeyWriter, concurrency int) error {
	var files []string
	var keys []string
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f.Mode().IsRegular() {
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return fmt.Errorf("could not get relative path for '%v': %v", path, err)
			}
			keys = append(keys, rel)
			files = append(files, path)
		}
		return nil
	})
	err := client.UploadFiles(files, keys, root, kw, concurrency)
	if err != nil {
		return fmt.Errorf("could not upload directory '%v': %v", dir, err)
	}
	return nil
}

//UploadFiles uploads a list of files concurrently.
func (client *DataClient) UploadFiles(files []string, keys []string, root string, kw KeyWriter, concurrency int) error {

	type item struct {
		filePath string
		key      string
		resCh    chan bool
		err      error
	}

	work := func(it *item) {
		it.err = client.UploadFile(it.filePath, it.key, root)
		it.resCh <- true
	}

	itemCh := make(chan *item, concurrency)
	go func() {
		defer close(itemCh)
		for i := 0; i < len(files); i++ {
			it := &item{
				filePath: files[i],
				key:      keys[i],
				resCh:    make(chan bool),
			}

			go work(it)  //create work
			itemCh <- it //send to fan-in thread for syncing results
		}
	}()

	//fan-in
	for it := range itemCh {
		<-it.resCh
		if it.err != nil {
			return fmt.Errorf("failed to upload '%v': %v", it.filePath, it.err)
		}

		err := kw.Write(it.filePath)
		if err != nil {
			return fmt.Errorf("failed to write key: %v", err)
		}
	}

	return nil
}

func (client *DataClient) ChunkedUpload(r io.Reader, kw KeyWriter, concurrency int, root string) (err error) {
	cr := chunker.New(r, chunker.Pol(0x3DA3358B4DC173))
	type result struct {
		err error
		k   string
	}

	type item struct {
		chunk []byte
		resCh chan *result
		err   error
	}

	work := func(it *item) {
		k := fmt.Sprintf("%x", sha256.Sum256(it.chunk)) //hash
		key := path.Join(root, k)
		exists, err := client.Has(key) //check existence
		if err != nil {
			it.resCh <- &result{fmt.Errorf("failed to check existence of '%x': %v", k, err), ""}
			return
		}

		if !exists {
			err = client.Upload(key, bytes.NewReader(it.chunk)) //if not exists put
			if err != nil {
				it.resCh <- &result{fmt.Errorf("failed to put chunk '%x': %v", k, err), ""}
				return
			}
		}

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

		err = kw.Write(res.k)
		if err != nil {
			return fmt.Errorf("failed to write key: %v", err)
		}
	}

	return nil
}

//DownloadFile downloads a single file.
func (client *DataClient) DownloadFile(key, outFile string) error {
	base := filepath.Dir(outFile)
	err := os.MkdirAll(base, DirectoryPermissions)
	if err != nil {
		return fmt.Errorf("failed to create path '%v': %v", base, err)
	}
	f, err := os.Create(outFile)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("failed to create local file '%v': %v", outFile, err)
	}

	svc := s3.New(client.Session)
	params := &s3.GetObjectInput{
		Bucket: aws.String(client.Bucket), // Required
		Key:    aws.String(key),           // Required
	}
	resp, err := svc.GetObject(params)

	if err != nil {
		return fmt.Errorf("failed to download '%v': %v", key, err)
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write output to '%v': %v", outFile, err)
	}

	return nil
}

//ListObjects lists all keys for a given root.
func (client *DataClient) ListObjects(root string) (keys []string, err error) {
	svc := s3.New(client.Session)

	params := &s3.ListObjectsInput{
		Bucket: aws.String(client.Bucket), // Required
		Prefix: aws.String(root),
	}
	resp, err := svc.ListObjects(params)

	if err != nil {
		return nil, fmt.Errorf("failed to list objects '%v': %v", root, err)
	}

	for _, object := range resp.Contents {
		keys = append(keys, aws.StringValue(object.Key))
	}

	return
}

//DownloadFiles concurrently downloads all files in a given s3 root path.
func (client *DataClient) DownloadFiles(root string, outDir string, kw KeyWriter, concurrency int, overwriteHandler func(string) bool) error {
	keys, err := client.ListObjects(root)
	if err != nil {
		return err
	}

	type item struct {
		key     string
		outFile string
		resCh   chan bool
		err     error
	}

	// build list of items with checks for overwrites if files exist
	items := make([]item, len(keys))
	totalItems := 0
	for i := 0; i < len(keys); i++ {
		it := item{
			key:     keys[i],
			outFile: path.Join(outDir, strings.Replace(keys[i], root+"/", "", 1)),
			resCh:   make(chan bool),
		}
		overwrite := true
		_, err = os.Stat(it.outFile)
		if err == nil || !os.IsNotExist(err) {
			overwrite = overwriteHandler(it.outFile)
		}
		if overwrite {
			items[totalItems] = it
			totalItems++
		}
	}

	work := func(it *item) {
		it.err = client.DownloadFile(it.key, it.outFile)
		it.resCh <- true
	}

	itemCh := make(chan *item, concurrency)
	go func() {
		defer close(itemCh)
		for i := 0; i < totalItems; i++ {
			go work(&items[i])  //create work
			itemCh <- &items[i] //send to fan-in thread for syncing results
		}
	}()

	//fan-in
	for it := range itemCh {
		<-it.resCh
		if it.err != nil {
			return fmt.Errorf("failed to download '%v': %v", it.key, it.err)
		}

		err := kw.Write(it.outFile)
		if err != nil {
			return fmt.Errorf("failed to write key: %v", err)
		}
	}

	return nil
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
