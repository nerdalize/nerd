package command

import (
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//Dataset command
type Dataset struct {
	*command
}

//DatasetFactory returns a factory method for the join command
func DatasetFactory() (cli.Command, error) {
	cmd := &Dataset{
		command: &command{
			help:     `upload and download data for tasks to use`,
			synopsis: "upload and download data for tasks to use",
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
	return errShowHelp
}
