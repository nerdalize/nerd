package command

import (
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

// Plan command
type Plan struct {
	*command
}

var synopsisPlan = "Set and list your plans."
var helpPlan = "A plan needs to be set to start using compute power."

// PlanFactory returns a factory method for the plan command
func PlanFactory() (cli.Command, error) {
	comm, err := newCommand("nerd plan <subcommand>", synopsisPlan, helpPlan, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &Project{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Plan) DoRun(args []string) (err error) {
	return errShowHelp("Not enough arguments, see below for usage.")
}
