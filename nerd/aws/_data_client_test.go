package aws

// import (
// 	"bytes"
// 	"errors"
// 	"io"
// 	"io/ioutil"
// 	"math/rand"
// 	"strings"
// 	"sync"
// 	"testing"
// 	"time"
//
// 	"github.com/aws/aws-sdk-go/aws"
// 	"github.com/aws/aws-sdk-go/aws/awserr"
// 	"github.com/aws/aws-sdk-go/aws/request"
// 	"github.com/aws/aws-sdk-go/aws/session"
// 	"github.com/aws/aws-sdk-go/service/s3"
// 	"github.com/nerdalize/nerd/nerd/data"
// )
//
// const KiB = 1024
// const MiB = KiB * 1024
//
// func newClient(t *testing.T, handler func(r *request.Request)) *DataClient {
// 	svc := s3.New(session.New())
// 	svc.Handlers.Clear()
// 	svc.Handlers.Send.PushBack(handler)
// 	return &DataClient{
// 		Service: svc,
// 		DataClientConfig: &DataClientConfig{
// 			Bucket: "test",
// 		},
// 	}
// }
//
// func randr(size int64, seed int64) io.Reader {
// 	if seed == 0 {
// 		seed = time.Now().UnixNano()
// 	}
//
// 	return io.LimitReader(rand.New(rand.NewSource(seed)), size)
// }
//
// func randb(size int64, seed int64) []byte {
// 	b, err := ioutil.ReadAll(randr(size, seed))
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	return b
// }
//
// type closingBuffer struct {
// 	*bytes.Reader
// }
//
// func (cb *closingBuffer) Close() error {
// 	//we don't actually have to do anything here, since the buffer is just some data in memory
// 	return nil
// }
//
// type dataStore struct {
// 	*sync.Mutex
// 	M map[string][]byte
// }
//
// func newMetadata() *data.Metadata {
// 	header := &data.MetadataHeader{
// 		Size:    0,
// 		Created: time.Now(),
// 		Updated: time.Now(),
// 	}
// 	return data.NewMetadata(header, data.NewBufferedKeyReadWiter())
// }
//
// //TestDownload tests Upload and its retry mechanism.
// func TestUpload(t *testing.T) {
// 	counter := 0
// 	cl := newClient(t, func(r *request.Request) {
// 		if counter == 0 {
// 			r.Error = errors.New("test")
// 		}
// 		counter += 1
// 	})
// 	err := cl.Upload("test", strings.NewReader(""))
// 	if err != nil {
// 		t.Errorf("Unexpected error: %v", err)
// 	}
// }
//
// //TestDownload tests Download and its retry mechanism.
// func TestDownload(t *testing.T) {
// 	counter := 0
// 	cl := newClient(t, func(r *request.Request) {
// 		if counter == 0 {
// 			r.Error = errors.New("test")
// 		}
// 		counter += 1
// 	})
// 	_, err := cl.Download("test")
// 	if err != nil {
// 		t.Errorf("Unexpected error: %v", err)
// 	}
// }
//
// //TestUpDown tests up/download end-to-end with a mocked S3 interface.
// func TestUpDown(t *testing.T) {
// 	progressCh := make(chan int64)
// 	go func() {
// 		for _ = range progressCh {
// 		}
// 	}()
//
// 	metadata := newMetadata()
// 	t.Run("all-exist", func(t *testing.T) {
// 		cl := newClient(t, func(r *request.Request) {
// 		})
// 		err := cl.ChunkedUpload(randr(5*MiB, 0), metadata, 10, "", progressCh)
// 		if err != nil {
// 			t.Errorf("Unexpected error: %v", err)
// 		}
// 	})
//
// 	metadata = newMetadata()
// 	input := randb(5*MiB, 0)
// 	ds := &dataStore{
// 		Mutex: new(sync.Mutex),
// 		M:     make(map[string][]byte),
// 	}
// 	t.Run("upload", func(t *testing.T) {
// 		upload(t, metadata, input, ds, progressCh)
// 	})
//
// 	var output = bytes.NewBuffer([]byte{})
// 	t.Run("download", func(t *testing.T) {
// 		download(t, metadata, output, ds, progressCh)
// 	})
//
// 	if !bytes.Equal(output.Bytes(), input) {
// 		t.Error("downloaded data should be equal to input")
// 	}
// }
//
// func upload(t *testing.T, metadata *data.Metadata, input []byte, ds *dataStore, progressCh chan int64) {
// 	cl := newClient(t, func(r *request.Request) {
// 		if r.Operation.Name == "HeadObject" {
// 			r.Error = awserr.New(s3.ErrCodeNoSuchKey, "", nil)
// 		} else {
// 			params := r.Params.(*s3.PutObjectInput)
// 			d, err := ioutil.ReadAll(params.Body)
// 			if err != nil {
// 				t.Errorf("Unexpected error: %v", err)
// 			}
// 			ds.Lock()
// 			ds.M[aws.StringValue(params.Key)] = d
// 			ds.Unlock()
// 		}
// 	})
// 	err := cl.ChunkedUpload(bytes.NewReader(input), metadata, 10, "", progressCh)
// 	if err != nil {
// 		t.Errorf("Unexpected error: %v", err)
// 	}
// }
//
// func download(t *testing.T, metadata *data.Metadata, output io.Writer, ds *dataStore, progressCh chan int64) {
// 	cl := newClient(t, func(r *request.Request) {
// 		params := r.Params.(*s3.GetObjectInput)
// 		d := ds.M[aws.StringValue(params.Key)]
// 		data := r.Data.(*s3.GetObjectOutput)
// 		data.Body = &closingBuffer{bytes.NewReader(d)}
// 	})
// 	err := cl.ChunkedDownload(metadata, output, 10, "", progressCh)
// 	if err != nil {
// 		t.Errorf("Unexpected error: %v", err)
// 	}
// }
