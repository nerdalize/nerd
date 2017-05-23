package command

import (
	"context"
	"fmt"
	"os"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	"github.com/nerdalize/nerd/nerd/aws"
	v1datatransfer "github.com/nerdalize/nerd/nerd/service/datatransfer/v1"
	"github.com/pkg/errors"
)

const (
	//DatasetFilename is the filename of the file that contains the dataset ID in the data folder.
	DatasetFilename = ".dataset"
	//DatasetPermissions are the permissions for DatasetFilename
	DatasetPermissions = 0644
	//UploadConcurrency is the amount of concurrent upload threads.
	UploadConcurrency = 64
)

//UploadOpts describes command options
type UploadOpts struct {
	Tag string `long:"tag" default:"" default-mask:"" description:"use a tag to logically group datasets"`
}

//Upload command
type Upload struct {
	*command

	opts *UploadOpts
}

//DatasetUploadFactory returns a factory method for the join command
func DatasetUploadFactory() (cli.Command, error) {
	opts := &UploadOpts{}
	comm, err := newCommand("nerd upload <path>", "upload data to the cloud and create a new dataset", "", opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &Upload{
		command: comm,
		opts:    opts,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Upload) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	dataPath := args[0]

	fi, err := os.Stat(dataPath)
	if err != nil {
		return errors.Errorf("argument '%v' is not a valid file or directory", dataPath)
	} else if !fi.IsDir() {
		return errors.Errorf("provided path '%s' is not a directory", dataPath)
	}

	// Clients
	batchclient, err := NewClient(cmd.ui, cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return err
	}
	ss, err := cmd.session.Read()
	if err != nil {
		return err
	}
	dataOps, err := aws.NewDataClient(
		aws.NewNerdalizeCredentials(batchclient, ss.Project.Name),
		ss.Project.AWSRegion,
	)
	if err != nil {
		return errors.Wrap(err, "could not create aws dataops client")
	}
	progressCh := make(chan int64)
	uploadConf := v1datatransfer.UploadConfig{
		BatchClient: batchclient,
		DataOps:     dataOps,
		LocalDir:    dataPath,
		ProjectID:   ss.Project.Name,
		Tag:         cmd.opts.Tag,
		Concurrency: 64,
		ProgressCh:  progressCh,
	}
	progressBarDoneCh := make(chan struct{})
	size, err := v1datatransfer.GetLocalDatasetSize(context.Background(), dataPath)
	if err != nil {
		return errors.Wrap(err, "failed to get dataset size")
	}
	go ProgressBar(cmd.outputter.ErrW(), size, progressCh, progressBarDoneCh)

	dataset, err := v1datatransfer.Upload(context.Background(), uploadConf)
	if err != nil {
		return errors.Wrap(err, "failed to upload dataset")
	}
	<-progressBarDoneCh
	tmpl := "$.DatasetID"
	jsonTmpl := "{\"dataset_id\":\"{{$.DatasetID}}}\"}"
	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.TmplDecorator(dataset, tmpl),
		format.OutputTypeRaw:    format.TmplDecorator(dataset, tmpl),
		format.OutputTypeJSON:   format.TmplDecorator(dataset, jsonTmpl),
	})
	return nil
}
