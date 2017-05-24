package command

import (
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//Dataset command
type Dataset struct {
	*command
}

//DatasetFactory returns a factory method for the join command
func DatasetFactory() (cli.Command, error) {
	comm, err := newCommand("nerd dataset <subcommand>", "upload and download datasets for tasks to use", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &Dataset{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Dataset) DoRun(args []string) (err error) {
	return errShowHelp
}
