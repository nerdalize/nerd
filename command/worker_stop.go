package command

import (
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//WorkerStop command
type WorkerStop struct {
	*command
}

//WorkerStopFactory returns a factory method for the join command
func WorkerStopFactory() (cli.Command, error) {
	comm, err := newCommand("nerd worker stop", "stop a worker from providing compute capacity", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &WorkerStop{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkerStop) DoRun(args []string) (err error) {
	return fmt.Errorf("not yet implemented")
}
