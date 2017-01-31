package data

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nerdalize/nerd/nerd"
)

type KeyWriter interface {
	Write(k string) error
}

type Client struct {
	Session *session.Session
}

func NewClient(awsCreds *credentials.Credentials) (*Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: awsCreds,
		Region:      aws.String("eu-west-1"),
	})
	if err != nil {
		return nil, fmt.Errorf("could not create AWS sessions: %v", err)
	}
	return &Client{
		Session: sess,
	}, nil
}

func (client *Client) UploadFile(filePath string, dataset string) error {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("could not open file '%v': %v", filePath, err)
	}
	svc := s3.New(client.Session)
	params := &s3.PutObjectInput{
		Bucket: aws.String(nerd.GetCurrentUser().AWSBucket),             // Required
		Key:    aws.String(path.Join(dataset, filepath.Base(filePath))), // Required
		Body:   file,
	}
	_, err = svc.PutObject(params)
	if err != nil {
		return fmt.Errorf("could not put file '%v': %v", filePath, err)
	}
	return nil
}

func (client *Client) UploadDir(dir string, dataset string) error {
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f.Mode().IsRegular() {
			return client.UploadFile(path, dataset)
		}
		return nil
	})
	return err
}

func (client *Client) UploadFiles(files []string, dataset string, kw KeyWriter, concurrency int) error {

	type item struct {
		filePath string
		resCh    chan bool
		err      error
	}

	work := func(it *item) {
		it.err = client.UploadFile(it.filePath, dataset)
		it.resCh <- true
	}

	itemCh := make(chan *item, concurrency)
	go func() {
		defer close(itemCh)
		for i := 0; i < len(files); i++ {
			it := &item{
				filePath: files[i],
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
