package v1datatransfer

import (
	"io"

	v1batch "github.com/nerdalize/nerd/nerd/client/batch/v1"
	v1data "github.com/nerdalize/nerd/nerd/service/datatransfer/v1/client"
	"github.com/pkg/errors"
)

//UploadConfig is the config for Upload operations
type UploadConfig struct {
	BatchClient *v1batch.Client
	DataOps     v1data.DataOps
	LocalDir    string
	ProjectID   string
	Tag         string
	Concurrency int
	ProgressCh  chan<- int64
}

//Upload uploads a dataset
func Upload(conf UploadConfig) (string, error) {
	ds, err := conf.BatchClient.CreateDataset(conf.ProjectID, conf.Tag)
	if err != nil {
		return "", errors.Wrap(err, "failed to create dataset")
	}
	dataClient := v1data.NewClient(conf.DataOps)
	up := &uploadProcess{
		batchClient:       conf.BatchClient,
		dataClient:        dataClient,
		dataset:           ds.DatasetSummary,
		heartbeatInterval: ds.HeartbeatInterval,
		localDir:          conf.LocalDir,
		concurrency:       conf.Concurrency,
		progressCh:        conf.ProgressCh,
	}
	return ds.DatasetID, up.start()
}

//GetLocalDatasetSize calculates the total size in bytes of the archived version of a directory on disk
func GetLocalDatasetSize(dataPath string) (int64, error) {
	type countResult struct {
		total int64
		err   error
	}
	doneCh := make(chan countResult)
	pr, pw := io.Pipe()
	go func() {
		total, err := countBytes(pr)
		doneCh <- countResult{total, err}
	}()

	err := tardir(dataPath, pw)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to tar '%s'", dataPath)
	}

	pw.Close()
	cr := <-doneCh
	if cr.err != nil {
		return 0, errors.Wrapf(err, "failed to count total disk size of '%v'", dataPath)
	}
	return cr.total, nil
}
