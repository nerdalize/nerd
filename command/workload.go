package command

import (
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//Workload command
type Workload struct {
	*command
}

//WorkloadFactory returns a factory method for the join command
func WorkloadFactory() (cli.Command, error) {
	comm, err := newCommand("nerd workload <subcommand>", "control compute capacity for working on tasks", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &Workload{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Workload) DoRun(args []string) (err error) {
	return errShowHelp
}
