package command

import (
	"strconv"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//TaskSuccess command
type TaskSuccess struct {
	*command
}

//TaskSuccessFactory returns a factory method for the join command
func TaskSuccessFactory() (cli.Command, error) {
	comm, err := newCommand("nerd task success <workload-id> <task-id> <run-token> <result>", "Mark a task run as having succeeded.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &TaskSuccess{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskSuccess) DoRun(args []string) (err error) {
	if len(args) < 4 {
		return errors.Wrap(errShowHelp("show help"), "Not enough arguments, see below for usage.")
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	taskID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return HandleError(errors.Wrap(err, "invalid task ID, must be a number"))
	}

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}

	projectID, err := ss.RequireProjectID()
	if err != nil {
		return HandleError(err)
	}

	out, err := bclient.SendRunSuccess(projectID, args[0], taskID, args[2], args[3], args[4])
	if err != nil {
		return HandleError(err)
	}

	cmd.outputter.Logger.Printf("Task Success: %v", out)
	return nil
}
