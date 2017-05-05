package command

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/aws"
	"github.com/nerdalize/nerd/nerd/conf"
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
	NerdOpts
	Tag string `long:"tag" default:"" default-mask:"" description:"use a tag to logically group datasets"`
}

//Upload command
type Upload struct {
	*command

	opts   *UploadOpts
	parser *flags.Parser
}

//DatasetUploadFactory returns a factory method for the join command
func DatasetUploadFactory() (cli.Command, error) {
	cmd := &Upload{
		command: &command{
			help:     "",
			synopsis: "upload data to the cloud and create a new dataset",
			parser:   flags.NewNamedParser("nerd upload <path>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &UploadOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

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
		HandleError(errors.Errorf("argument '%v' is not a valid file or directory", dataPath), cmd.opts.VerboseOutput)
	} else if !fi.IsDir() {
		HandleError(errors.Errorf("provided path '%s' is not a directory", dataPath), cmd.opts.VerboseOutput)
	}

	// Config
	config, err := conf.Read()
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
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

	err = v1datatransfer.Upload(v1datatransfer.UploadConfig{
		BatchClient: batchclient,
		DataOps:     dataOps,
		LocalDir:    dataPath,
		ProjectID:   config.CurrentProject.Name,
		Tag:         cmd.opts.Tag,
		Concurrency: 64,
		ProgressCh:  nil,
	})
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}
	return nil
}
