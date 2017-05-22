package command

import (
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//WorkerStart command
type WorkerStart struct {
	*command
}

//WorkerStartFactory returns a factory method for the join command
func WorkerStartFactory() (cli.Command, error) {
	comm, err := newCommand("nerd worker start", "provision a new worker to provide compute", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &WorkerStart{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkerStart) DoRun(args []string) (err error) {
	return fmt.Errorf("not yet implemented")
}
