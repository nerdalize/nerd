package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/aws"
	"github.com/nerdalize/nerd/nerd/conf"
	v1datatransfer "github.com/nerdalize/nerd/nerd/service/datatransfer/v1"
	"github.com/pkg/errors"
)

const (
	//OutputDirPermissions are the output directory's permissions.
	OutputDirPermissions = 0755
	//DownloadConcurrency is the amount of concurrent download threads.
	DownloadConcurrency = 64
	//DatasetPrefix is the prefix of each dataset ID.
	DatasetPrefix = "d-"
	TagPrefix     = "tag-"
)

//DownloadOpts describes command options
type DownloadOpts struct {
	NerdOpts
	AlwaysOverwrite bool `long:"always-overwrite" default-mask:"false" description:"always overwrite files when they already exist"`
}

//Download command
type Download struct {
	*command

	opts   *DownloadOpts
	parser *flags.Parser
}

//DatasetDownloadFactory returns a factory method for the join command
func DatasetDownloadFactory() (cli.Command, error) {
	cmd := &Download{
		command: &command{
			help:     "",
			synopsis: "download data from the cloud to a local directory",
			parser:   flags.NewNamedParser("nerd dataset download <dataset> <output-dir>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &DownloadOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Download) DoRun(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	config, err := conf.Read()
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	downloadObject := args[0]
	outputDir := args[1]

	// Folder create and check
	fi, err := os.Stat(outputDir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, OutputDirPermissions)
		if err != nil {
			HandleError(errors.Errorf("The provided path '%s' does not exist and could not be created.", outputDir), cmd.opts.VerboseOutput)
		}
		fi, err = os.Stat(outputDir)
	}
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	} else if !fi.IsDir() {
		HandleError(errors.Errorf("The provided path '%s' is not a directory", outputDir), cmd.opts.VerboseOutput)
	}

	// Clients
	batchclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}
	dataOps, err := aws.NewDataClient(
		aws.NewNerdalizeCredentials(batchclient, config.CurrentProject.Name),
		config.CurrentProject.AWSRegion,
	)
	if err != nil {
		HandleError(errors.Wrap(err, "could not create aws dataops client"), cmd.opts.VerboseOutput)
	}

	// Gather dataset IDs
	var datasetIDs []string
	if !strings.HasPrefix(downloadObject, TagPrefix) {
		datasetIDs = append(datasetIDs, downloadObject)
	} else {
		datasets, err := batchclient.ListDatasets(config.CurrentProject.Name, downloadObject)
		if err != nil {
			HandleError(err, cmd.opts.VerboseOutput)
		}
		datasetIDs = make([]string, len(datasets.Datasets))
		for i, dataset := range datasets.Datasets {
			datasetIDs[i] = dataset.DatasetID
		}
	}

	for _, datasetID := range datasetIDs {
		logrus.Infof("Downloading dataset with ID '%v'", datasetID)
		downloadConf := v1datatransfer.DownloadConfig{
			BatchClient: batchclient,
			DataOps:     dataOps,
			LocalDir:    outputDir,
			ProjectID:   config.CurrentProject.Name,
			DatasetID:   datasetID,
			Concurrency: 64,
		}
		if !cmd.opts.JSONOutput { // show progress bar
			progressCh := make(chan int64)
			progressBarDoneCh := make(chan struct{})
			size, err := v1datatransfer.GetRemoteDatasetSize(batchclient, dataOps, config.CurrentProject.Name, datasetID)
			if err != nil {
				HandleError(err, cmd.opts.VerboseOutput)
			}
			go ProgressBar(size, progressCh, progressBarDoneCh)
			downloadConf.ProgressCh = progressCh
			err = v1datatransfer.DownloadBlocking(downloadConf)
			if err != nil {
				HandleError(errors.Wrapf(err, "failed to download dataset '%v'", datasetID), cmd.opts.VerboseOutput)
			}
			<-progressBarDoneCh
		} else { //do not show progress bar
			err = v1datatransfer.DownloadBlocking(downloadConf)
			if err != nil {
				HandleError(errors.Wrapf(err, "failed to download dataset '%v'", datasetID), cmd.opts.VerboseOutput)
			}
		}
	}

	return nil
}
