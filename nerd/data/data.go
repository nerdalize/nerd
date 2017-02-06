package data

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

//Client holds a reference to an AWS session
type Client struct {
	Session *session.Session
}

//NewClient creates a new data client that is capable of uploading and downloading (multiple) files.
func NewClient(awsCreds *credentials.Credentials) (*Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: awsCreds,
		Region:      aws.String(nerd.GetCurrentUser().Region),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create AWS sessions: %v", err)
	}
	return &Client{
		Session: sess,
	}, nil
}

//UploadFile uploads a single file.
func (client *Client) UploadFile(filePath string, key string, dataset string) error {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("could not open file '%v': %v", filePath, err)
	}
	svc := s3.New(client.Session)
	params := &s3.PutObjectInput{
		Bucket: aws.String(nerd.GetCurrentUser().AWSBucket), // Required
		Key:    aws.String(path.Join(dataset, key)),         // Required
		Body:   file,
	}
	_, err = svc.PutObject(params)
	if err != nil {
		return fmt.Errorf("could not put file '%v': %v", filePath, err)
	}
	return nil
}

//UploadDir uploads every single file in the directory and all its subdirectories.
func (client *Client) UploadDir(dir string, dataset string, kw KeyWriter, concurrency int) error {
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
	err := client.UploadFiles(files, keys, dataset, kw, concurrency)
	if err != nil {
		return fmt.Errorf("could not upload directory '%v': %v", dir, err)
	}
	return nil
}

//UploadFiles uploads a list of files concurrently.
func (client *Client) UploadFiles(files []string, keys []string, dataset string, kw KeyWriter, concurrency int) error {

	type item struct {
		filePath string
		key      string
		resCh    chan bool
		err      error
	}

	work := func(it *item) {
		it.err = client.UploadFile(it.filePath, it.key, dataset)
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
func (client *Client) DownloadFile(key string, outDir string) error {
	//strip the "dataset" prefix
	stripped := strings.Join(strings.Split(key, "/")[1:], "/")
	base := filepath.Dir(path.Join(outDir, stripped))
	err := os.MkdirAll(base, DirectoryPermissions)
	if err != nil {
		return fmt.Errorf("failed to create path '%v': %v", base, err)
	}
	outFile, err := os.Create(path.Join(outDir, stripped))
	defer outFile.Close()
	if err != nil {
		return fmt.Errorf("failed to create local file '%v': %v", path.Join(outDir, key), err)
	}

	svc := s3.New(client.Session)
	params := &s3.GetObjectInput{
		Bucket: aws.String(nerd.GetCurrentUser().AWSBucket), // Required
		Key:    aws.String(key),                             // Required
	}
	resp, err := svc.GetObject(params)

	if err != nil {
		return fmt.Errorf("failed to download '%v': %v", key, err)
	}

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write output to '%v': %v", path.Join(outDir, key), err)
	}

	return nil
}

//ListDataset lists all keys for a given dataset.
func (client *Client) ListDataset(dataset string) (keys []string, err error) {
	svc := s3.New(client.Session)

	params := &s3.ListObjectsInput{
		Bucket: aws.String(nerd.GetCurrentUser().AWSBucket), // Required
		Prefix: aws.String(dataset),
	}
	resp, err := svc.ListObjects(params)

	if err != nil {
		return nil, fmt.Errorf("failed to list dataset '%v': %v", dataset, err)
	}

	for _, object := range resp.Contents {
		keys = append(keys, aws.StringValue(object.Key))
	}

	return
}

//DownloadFiles concurrently downloads all files in a given dataset.
func (client *Client) DownloadFiles(dataset string, outDir string, kw KeyWriter, concurrency int) error {
	keys, err := client.ListDataset(dataset)
	if err != nil {
		return err
	}

	type item struct {
		key    string
		outDir string
		resCh  chan bool
		err    error
	}

	work := func(it *item) {
		it.err = client.DownloadFile(it.key, it.outDir)
		it.resCh <- true
	}

	itemCh := make(chan *item, concurrency)
	go func() {
		defer close(itemCh)
		for i := 0; i < len(keys); i++ {
			it := &item{
				key:    keys[i],
				outDir: outDir,
				resCh:  make(chan bool),
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

		stripped := strings.Join(strings.Split(it.key, "/")[1:], "/")
		err := kw.Write(path.Join(outDir, stripped))
		if err != nil {
			return fmt.Errorf("failed to write key: %v", err)
		}
	}

	return nil
}
