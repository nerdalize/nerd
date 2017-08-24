package command

import (
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//Worker command
type Worker struct {
	*command
}

//WorkerFactory returns a factory method for the join command
func WorkerFactory() (cli.Command, error) {
	comm, err := newCommand("nerd worker <subcommand>", "Control individual compute processes.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &Worker{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Worker) DoRun(args []string) (err error) {
	return errShowHelp("Not enough arguments, see below for usage.")
}
