package cmd

import (
	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//Dataset command
type Dataset struct {
	*command
}

//DatasetFactory creates the command
func DatasetFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &Dataset{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, nil, flags.None, "nerd dataset")

	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *Dataset) Execute(args []string) (err error) { return errShowHelp("") }

// Description returns long-form help text
func (cmd *Dataset) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *Dataset) Synopsis() string {
	return "Group of commands used to manage datasets. A dataset is a collection of files, like a folder on a computer. They can be used as input for a job and, when an application creates output files, these can be automatically stored in a new dataset."
}

// Usage shows usage
func (cmd *Dataset) Usage() string { return "nerd dataset <subcommand>" }
