package v1datatransfer

import (
	v1batch "github.com/nerdalize/nerd/nerd/client/batch/v1"
	v1data "github.com/nerdalize/nerd/nerd/service/datatransfer/v1/client"
	"github.com/pkg/errors"
)

type UploadConfig struct {
	BatchClient *v1batch.Client
	DataOps     v1data.DataOps
	LocalDir    string
	ProjectID   string
	Tag         string
	Concurrency int
	ProgressCh  chan int64
}

func Upload(conf UploadConfig) error {
	ds, err := conf.BatchClient.CreateDataset(conf.ProjectID, conf.Tag)
	if err != nil {
		return errors.Wrap(err, "failed to create dataset")
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
	return up.start()
}
