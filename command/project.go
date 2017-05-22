package command

import (
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//Project command
type Project struct {
	*command
}

//ProjectFactory returns a factory method for the join command
func ProjectFactory() (cli.Command, error) {
	comm, err := newCommand("nerd project <subcommand>", "set and list projects", "", nil)
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
func (cmd *Project) DoRun(args []string) (err error) {
	return errShowHelp
}
