package command

import (
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//Upload command
type Dataset struct {
	*command
}

//DatasetUploadFactory returns a factory method for the join command
func DatasetFactory() (cli.Command, error) {
	cmd := &Dataset{
		command: &command{
			help:     `Upload and download datasets to and from Nerdalize Cloud Storage`,
			synopsis: "Upload and download datasets to and from Nerdalize Cloud Storage",
			parser:   flags.NewNamedParser("nerd dataset <subcommand>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},
	}

	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Dataset) DoRun(args []string) (err error) {
	return showHelpError
}
