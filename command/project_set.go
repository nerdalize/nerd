package command

import (
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

//ProjectSet command
type ProjectSet struct {
	*command
}

//ProjectSetFactory returns a factory method for the join command
func ProjectSetFactory() (cli.Command, error) {
	comm, err := newCommand("nerd project set", "Set current working project.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &ProjectSet{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *ProjectSet) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return errors.Wrap(errShowHelp("show error"), "Not enough arguments, see below for usage.")
	}

	err = cmd.session.WriteProject(args[0], conf.DefaultAWSRegion)
	if err != nil {
		return HandleError(err)
	}

	return nil
}
