package command

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/aws"
	"github.com/pkg/errors"
)

const (
	OutputDirPermissions = 0755
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

//DownloadFactory returns a factory method for the join command
func DownloadFactory() func() (cmd cli.Command, err error) {
	cmd := &Download{
		command: &command{
			help:     "",
			synopsis: "Download a dataset from cloud storage",
			parser:   flags.NewNamedParser("nerd download <dataset> <output-dir>", flags.Default),
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

	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//DoRun is called by run and allows an error to be returned
func (cmd *Download) DoRun(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments, see --help")
	}
	dataset := args[0]
	outputDir := args[1]

	fi, err := os.Stat(outputDir)
	// create directory if it does not exist yet.
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

	nerdclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}
	ds, err := nerdclient.GetDataset(dataset)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	client, err := aws.NewDataClient(&aws.DataClientConfig{
		Credentials: aws.NewNerdalizeCredentials(nerdclient),
		Bucket:      ds.Bucket,
	})
	if err != nil {
		HandleError(errors.Wrap(err, "could not create data client"), cmd.opts.VerboseOutput)
	}

	overwriteHandler := OverwriteHandlerUserPrompt(cmd.ui)
	if cmd.opts.AlwaysOverwrite {
		overwriteHandler = AlwaysOverwriteHandler
	}
	err = client.DownloadFiles(ds.Root, outputDir, &stdoutkw{}, 64, overwriteHandler)
	if err != nil {
		HandleError(errors.Wrap(err, "could not download files"), cmd.opts.VerboseOutput)
	}

	return nil
}

//OverwriteHandlerUserPrompt is a handler that checks wether a file should be overwritten by asking the user over Stdin.
func OverwriteHandlerUserPrompt(ui cli.Ui) func(string) bool {
	return func(file string) bool {
		question := fmt.Sprintf("The file '%v' already exists. Do you want to overwrite it? [Y/n]", file)
		ans, err := ui.Ask(question)
		if err != nil {
			ui.Info(fmt.Sprintf("Failed to read your answer, '%v' will be skipped", file))
			return false
		}
		if ans == "n" {
			return false
		}
		return true
	}
}

//AlwaysOverwriteHandler is a handler that tells to always overwrite a file.
func AlwaysOverwriteHandler(file string) bool {
	return true
}
