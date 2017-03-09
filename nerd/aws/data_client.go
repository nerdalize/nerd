package aws

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nerdalize/nerd/nerd"
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

//UploadFile uploads a single file.
func (client *DataClient) UploadFile(filePath string, key string, root string) error {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("could not open file '%v': %v", filePath, err)
	}
	svc := s3.New(client.Session)
	params := &s3.PutObjectInput{
		Bucket: aws.String(client.Bucket),        // Required
		Key:    aws.String(path.Join(root, key)), // Required
		Body:   file,
	}
	_, err = svc.PutObject(params)
	if err != nil {
		return fmt.Errorf("could not put file '%v': %v", filePath, err)
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
func (client *DataClient) DownloadFiles(root string, outDir string, kw KeyWriter, concurrency int) error {
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

	work := func(it *item) {
		it.err = client.DownloadFile(it.key, it.outFile)
		it.resCh <- true
	}

	itemCh := make(chan *item, concurrency)
	go func() {
		defer close(itemCh)
		for i := 0; i < len(keys); i++ {
			it := &item{
				key:     keys[i],
				outFile: path.Join(outDir, strings.Replace(keys[i], root+"/", "", 1)),
				resCh:   make(chan bool),
			}

			go work(it)  //create work
			itemCh <- it //send to fan-in thread for syncing results
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
