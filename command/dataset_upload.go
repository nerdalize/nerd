package command

import (
	"context"
	"os"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	"github.com/nerdalize/nerd/nerd/aws"
	v1datatransfer "github.com/nerdalize/nerd/nerd/service/datatransfer/v1"
	"github.com/pkg/errors"
)

const (
	//UploadConcurrency is the amount of concurrent upload threads.
	UploadConcurrency = 64
)

//Upload command
type Upload struct {
	*command
}

//DatasetUploadFactory returns a factory method for the join command
func DatasetUploadFactory() (cli.Command, error) {
	comm, err := newCommand("nerd dataset upload <path>", "Upload data from a directory to the cloud and create a new dataset.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &Upload{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Upload) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return errShowHelp("Not enough arguments, see below for usage.")
	}

	dataPath := args[0]

	fi, err := os.Stat(dataPath)
	if err != nil {
		return errors.Errorf("argument '%v' is not a valid file or directory", dataPath)
	} else if !fi.IsDir() {
		return errors.Errorf("provided path '%s' is not a directory", dataPath)
	}

	// Clients
	batchclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return err
	}
	ss, err := cmd.session.Read()
	if err != nil {
		return err
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
		return errors.Wrap(err, "could not create aws dataops client")
	}
	progressCh := make(chan int64)
	progressBarDoneCh := make(chan struct{})
	size, err := v1datatransfer.GetLocalDatasetSize(context.Background(), dataPath)
	if err != nil {
		return HandleError(err)
	}
	go ProgressBar(cmd.outputter.ErrW(), size, progressCh, progressBarDoneCh)
	uploadConf := v1datatransfer.UploadConfig{
		BatchClient: batchclient,
		DataOps:     dataOps,
		LocalDir:    dataPath,
		ProjectID:   projectID,
		Concurrency: UploadConcurrency,
		ProgressCh:  progressCh,
	}

	dataset, err := v1datatransfer.Upload(context.Background(), uploadConf)
	if err != nil {
		return HandleError(err)
	}
	<-progressBarDoneCh
	tmpl := "Upload complete! New dataset ID: {{$.DatasetID}}\n"
	jsonTmpl := "{\"dataset_id\":\"{{$.DatasetID}}}\"}"
	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTmplDecorator(dataset, tmpl),
		format.OutputTypeRaw:    format.NewTmplDecorator(dataset, tmpl),
		format.OutputTypeJSON:   format.NewTmplDecorator(dataset, jsonTmpl),
	})
	return nil
}
