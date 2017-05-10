package v1datatransfer

import (
	"context"
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
func Upload(ctx context.Context, conf UploadConfig) (string, error) {
	if conf.ProgressCh != nil {
		defer close(conf.ProgressCh)
	}
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
	return ds.DatasetID, up.start(ctx)
}

//GetLocalDatasetSize calculates the total size in bytes of the archived version of a directory on disk
func GetLocalDatasetSize(ctx context.Context, dataPath string) (int64, error) {
	type countResult struct {
		total int64
		err   error
	}
	doneCh := make(chan countResult)
	pr, pw := io.Pipe()
	go func() {
		total, err := countBytes(ctx, pr)
		doneCh <- countResult{total, err}
	}()

	err := tardir(ctx, dataPath, pw)
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
