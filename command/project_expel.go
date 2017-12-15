package command

import (
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//ProjectExpel command
type ProjectExpel struct {
	*command
}

//ProjectExpelFactory returns a factory method for the join command
func ProjectExpelFactory() (cli.Command, error) {
	comm, err := newCommand("nerd project expel", "Move the current project away from its current cluster.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &ProjectExpel{
		command: comm,
	}

	cmd.runFunc = cmd.DoRun
	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *ProjectExpel) DoRun(args []string) (err error) {
	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}

	projectID, err := ss.RequireProjectID()
	if err != nil {
		return HandleError(err)
	}

	_, err = bclient.ExpelProject(projectID)
	if err != nil {
		return HandleError(err)
	}

	cmd.outputter.Logger.Printf("Successfully removed project placement")
	return nil
}
