package command

import (
	"context"
	"fmt"
	"os"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/aws"
	v1datatransfer "github.com/nerdalize/nerd/nerd/service/datatransfer/v1"
	"github.com/pkg/errors"
)

const (
	//OutputDirPermissions are the output directory's permissions.
	OutputDirPermissions = 0755
	//DownloadConcurrency is the amount of concurrent download threads.
	DownloadConcurrency = 10
)

//Download command
type Download struct {
	*command
}

//DatasetDownloadFactory returns a factory method for the join command
func DatasetDownloadFactory() (cli.Command, error) {
	comm, err := newCommand("nerd dataset download <dataset-id> <output-dir>", "download data from the cloud to a local directory", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &Download{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Download) DoRun(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	datasetID := args[0]
	outputDir := args[1]

	// Folder create and check
	fi, err := os.Stat(outputDir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, OutputDirPermissions)
		if err != nil {
			return HandleError(errors.Errorf("The provided path '%s' does not exist and could not be created.", outputDir))
		}
		fi, err = os.Stat(outputDir)
	}
	if err != nil {
		return HandleError(err)
	} else if !fi.IsDir() {
		return HandleError(errors.Errorf("The provided path '%s' is not a directory", outputDir))
	}

	// Clients
	batchclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}
	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}

	projectID, err := ss.RequireProjectID()
	if err != nil {
		return HandleError(err)
	}

	dataOps, err := aws.NewDataClient(
		aws.NewNerdalizeCredentials(batchclient, projectID),
		ss.Project.AWSRegion,
	)
	if err != nil {
		return HandleError(errors.Wrap(err, "could not create aws dataops client"))
	}

	cmd.outputter.Logger.Printf("Downloading dataset with ID '%v'", datasetID)
	downloadConf := v1datatransfer.DownloadConfig{
		BatchClient: batchclient,
		DataOps:     dataOps,
		LocalDir:    outputDir,
		ProjectID:   projectID,
		DatasetID:   datasetID,
		Concurrency: DownloadConcurrency,
	}

	progressCh := make(chan int64)
	progressBarDoneCh := make(chan struct{})
	var size int64
	size, err = v1datatransfer.GetRemoteDatasetSize(context.Background(), batchclient, dataOps, projectID, datasetID)
	if err != nil {
		return HandleError(err)
	}

	go ProgressBar(cmd.outputter.ErrW(), size, progressCh, progressBarDoneCh)
	downloadConf.ProgressCh = progressCh
	err = v1datatransfer.Download(context.Background(), downloadConf)
	if err != nil {
		return HandleError(errors.Wrapf(err, "failed to download dataset '%v'", datasetID))
	}

	<-progressBarDoneCh
	return nil
}
