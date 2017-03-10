package command

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/aws"
	"github.com/pkg/errors"
)

//DownloadOpts describes command options
type DownloadOpts struct {
	NerdOpts
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
			synopsis: "fetch the output of a task from cloud storage",
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
	if err != nil {
		return fmt.Errorf("failed to inspect '%s': %v", outputDir, err)
	} else if !fi.IsDir() {
		return fmt.Errorf("provided path '%s' is not a directory", outputDir)
	}

	nerdclient, err := NewClient(cmd.ui)
	if err != nil {
		return HandleError(HandleClientError(err, cmd.opts.VerboseOutput), cmd.opts.VerboseOutput)
	}
	ds, err := nerdclient.GetDataset(dataset)
	if err != nil {
		return errors.Wrapf(err, "failed to get dataset information for dataset %v", dataset)
	}

	client, err := aws.NewDataClient(&aws.DataClientConfig{
		Credentials: aws.NewNerdalizeCredentials(nerdclient),
		Bucket:      ds.Bucket,
	})
	if err != nil {
		return fmt.Errorf("could not create data client: %v", err)
	}

	err = client.DownloadFiles(ds.Root, outputDir, &stdoutkw{}, 64)
	if err != nil {
		return fmt.Errorf("could not download files: %v", err)
	}

	return nil
}
