package command

import (
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

// PlanRemove command
type PlanRemove struct {
	*command
}

// PlanRemoveFactory returns a factory method for the plan remove command
func PlanRemoveFactory() (cli.Command, error) {
	comm, err := newCommand("nerd plan remove <name>", "Remove a plan from your project.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &PlanRemove{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *PlanRemove) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return errShowHelp("Not enough arguments, see below for usage.")
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}
	_, err = ss.RequireProjectID()
	if err != nil {
		return HandleError(err)
	}

	_, err = bclient.RemovePlan(ss.Project.Name, args[0])
	if err != nil {
		return HandleError(err)
	}

	cmd.outputter.Logger.Printf("Plan successfully removed")
	return nil
}
