package command

import (
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//Queue command
type Queue struct {
	*command
}

//QueueFactory returns a factory method for the join command
func QueueFactory() (cli.Command, error) {
	comm, err := newCommand("nerd queue <subcommand>", "setup queues that transport tasks to workers", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &Queue{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Queue) DoRun(args []string) (err error) {
	return errShowHelp
}
