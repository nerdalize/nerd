package v1datatransfer

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
	v1data "github.com/nerdalize/nerd/nerd/service/datatransfer/v1/client"
)

func testfile(dir string, name string, size, seed int64, t interface {
	Fatalf(format string, args ...interface{})
}) (os.FileInfo, []byte) {
	path := filepath.Join(dir, name)
	err := os.MkdirAll(filepath.Dir(path), 0777)
	if err != nil {
		t.Fatalf("failed to create file dir for '%v': %v", path, err)
	}

	data := randb(size, seed)
	err = ioutil.WriteFile(path, data, 0666)
	if err != nil {
		t.Fatalf("failed to write file '%v': %v", path, err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("failed to stat test file '%v': %v", path, err)
	}

	return fi, data
}

type clientUpload struct{}

func (c *clientUpload) SendUploadHeartbeat(projectID, datasetID string) (output *v1payload.SendUploadHeartbeatOutput, err error) {
	return output, err
}
func (c *clientUpload) SendUploadSuccess(projectID, datasetID string) (output *v1payload.SendUploadSuccessOutput, err error) {
	return output, err
}

func TestUploadContext(t *testing.T) {
	baseNum := runtime.NumGoroutine()
	dir, err := ioutil.TempDir("", "test_upload_context")
	if err != nil {
		t.Fatalf("failed to setup tempdir: %v", err)
	}
	name := "test"
	testfile(dir, name, 10*MiB, 42, t)

	dp := &uploadProcess{
		batchClient:       &clientUpload{},
		dataClient:        v1data.NewClient(&blockingOps{}),
		dataset:           v1payload.DatasetSummary{},
		heartbeatInterval: time.Second * 10,
		localDir:          dir,
		concurrency:       5,
		progressCh:        nil,
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		dp.start(ctx)
	}()
	time.Sleep(time.Second * 5)
	expected := baseNum + 13
	if runtime.NumGoroutine() != expected {
		t.Fatalf("expected %v goroutines, got: %v", expected, runtime.NumGoroutine())
	}
	cancel()
	time.Sleep(time.Second * 5)
	if runtime.NumGoroutine() != baseNum {
		t.Fatalf("expected %v goroutines, got: %v", baseNum, runtime.NumGoroutine())
	}
}
