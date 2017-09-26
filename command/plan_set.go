package command

import (
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

// PlanSetOpts describes the options to the PlanAdd command
type PlanSetOpts struct {
	Username string `long:"username" default:"" default-mask:"" description:"Username for Docker registry authentication"`
	Password string `long:"password" default:"" default-mask:"" description:"Password for Docker registry authentication"`
	Type     string `long:"type" default:"opaque" default-mask:"" description:"Type of plan to display"`
}

// PlanSet command
type PlanSet struct {
	*command
	opts *PlanSetOpts
}

// PlanSetFactory returns a factory method for the plan command
func PlanSetFactory() (cli.Command, error) {
	opts := &PlanSetOpts{}
	comm, err := newCommand("nerd plan set", "", "", opts)
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
func (cmd *PlanSet) DoRun(args []string) (err error) {
	return errShowHelp("Not enough arguments, see below for usage.")
}
