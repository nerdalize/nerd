package v1datatransfer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
	v1data "github.com/nerdalize/nerd/nerd/service/datatransfer/v1/client"
	"github.com/pkg/errors"
)

const (
	KiB = 1024
	MiB = KiB * 1024
)

var (
	DownloadFailKey v1data.Key = [sha256.Size]byte{0, 1, 2}
	ErrDownloadFail            = fmt.Errorf("download failed")
)

func randr(size int64, seed int64) io.Reader {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	return io.LimitReader(rand.New(rand.NewSource(seed)), size)
}

func randb(size int64, seed int64) []byte {
	b, err := ioutil.ReadAll(randr(size, seed))
	if err != nil {
		panic(err)
	}

	return b
}

type fakeDataOps struct {
	data map[string][]byte
	m    *sync.Mutex
}

func (f *fakeDataOps) Upload(ctx context.Context, bucket, key string, body io.ReadSeeker) error {
	f.m.Lock()
	defer f.m.Unlock()

	dat, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	f.data[bucket+"/"+key] = dat
	return nil
}
func (f *fakeDataOps) Download(ctx context.Context, bucket, key string) (body io.ReadCloser, err error) {
	f.m.Lock()
	defer f.m.Unlock()
	if strings.Contains(key, DownloadFailKey.ToString()) {
		return nil, ErrDownloadFail
	}
	return ioutil.NopCloser(bytes.NewReader(f.data[bucket+"/"+key])), nil
}
func (f *fakeDataOps) Exists(ctx context.Context, bucket, key string) (exists bool, err error) {
	f.m.Lock()
	defer f.m.Unlock()
	_, ok := f.data[bucket+"/"+key]
	return ok, nil
}

func newClient() *v1data.Client {
	return v1data.NewClient(&fakeDataOps{
		data: make(map[string][]byte),
		m:    &sync.Mutex{},
	})
}

//BufferedKeyReadWriter contains an internal buffer of Keys which could be read and written to.
type BufferedKeyReadWriter struct {
	*sync.Mutex
	pos int
	M   map[v1data.Key]struct{}
	L   []v1data.Key
}

//NewBufferedKeyReadWiter creates a new BufferedKeyReadWriter.
func NewBufferedKeyReadWiter() *BufferedKeyReadWriter {
	return &BufferedKeyReadWriter{Mutex: &sync.Mutex{}, M: map[v1data.Key]struct{}{}}
}

//WriteKey writes a Key to the buffer.
func (kw *BufferedKeyReadWriter) WriteKey(k v1data.Key) error {
	kw.Lock()
	defer kw.Unlock()
	if _, ok := kw.M[k]; ok {
		return nil
	}

	kw.M[k] = struct{}{}
	kw.L = append(kw.L, k)
	return nil
}

//ReadKey reads a Key from the buffer and returns an io.EOF if no more Keys are available.
func (kw *BufferedKeyReadWriter) ReadKey() (k v1data.Key, err error) {
	kw.Lock()
	defer kw.Unlock()
	if kw.pos == len(kw.L) {
		return v1data.ZeroKey, io.EOF
	}

	k = kw.L[kw.pos]
	kw.pos = kw.pos + 1
	return k, nil
}

//TestUpDown tests up/download end-to-end with a mocked S3 backend.
func TestUpDown(t *testing.T) {
	cl := newClient()
	var input = randb(12*1024*1024, 0)
	var output = bytes.NewBuffer(nil)
	keyrw := NewBufferedKeyReadWiter()

	t.Run("upload", func(t *testing.T) {
		uploadChunks(context.Background(), cl, bytes.NewReader(input), keyrw, "bucket", "root", 10, nil)
	})

	t.Run("download", func(t *testing.T) {
		downloadChunks(context.Background(), cl, keyrw, output, 10, "bucket", "root", nil)
	})

	if !bytes.Equal(output.Bytes(), input) {
		t.Error("downloaded data should be equal to expected")
	}
}

func TestDownloadFail(t *testing.T) {
	cl := newClient()
	output := bytes.NewBuffer(nil)
	keyrw := NewBufferedKeyReadWiter()
	keyrw.WriteKey([sha256.Size]byte{0})
	keyrw.WriteKey(DownloadFailKey)
	keyrw.WriteKey([sha256.Size]byte{1})

	err := downloadChunks(context.Background(), cl, keyrw, output, 10, "bucket", "root", nil)
	if err == nil {
		t.Fatal("expected error but got none")
	}
	if errors.Cause(err) != ErrDownloadFail {
		t.Errorf("expected error cause to be %v, but was %v", errors.Cause(err), err)
	}
}

type blockingOps struct {
}

func (f *blockingOps) Upload(ctx context.Context, bucket, key string, body io.ReadSeeker) error {
	<-ctx.Done()
	return ctx.Err()
}
func (f *blockingOps) Download(ctx context.Context, bucket, key string) (body io.ReadCloser, err error) {
	if strings.Contains(key, "index") {
		return ioutil.NopCloser(strings.NewReader("01\n02\n03\n04\n05")), nil
	}
	<-ctx.Done()
	return nil, ctx.Err()
}
func (f *blockingOps) Exists(ctx context.Context, bucket, key string) (exists bool, err error) {
	<-ctx.Done()
	return false, ctx.Err()
}

func TestContext(t *testing.T) {
	baseNum := runtime.NumGoroutine()
	dp := &downloadProcess{
		dataClient:  v1data.NewClient(&blockingOps{}),
		dataset:     v1payload.DatasetSummary{},
		localDir:    "test",
		concurrency: 5,
		progressCh:  nil,
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		dp.start(ctx)
	}()
	time.Sleep(time.Second)
	expected := baseNum + 8
	if runtime.NumGoroutine() != expected {
		t.Fatalf("expected %v goroutines, got: %v", expected, runtime.NumGoroutine())
	}
	cancel()
	time.Sleep(time.Second)
	if runtime.NumGoroutine() != baseNum {
		t.Fatalf("expected %v goroutines, got: %v", baseNum, runtime.NumGoroutine())
	}
}
