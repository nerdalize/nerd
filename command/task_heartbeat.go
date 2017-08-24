package command

import (
	"strconv"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//TaskHeartbeat command
type TaskHeartbeat struct {
	*command
}

//TaskHeartbeatFactory returns a factory method for the join command
func TaskHeartbeatFactory() (cli.Command, error) {
	comm, err := newCommand("nerd task heartbeat <workload-id> <task-id> <run-token>", "Indicate that a task run is still in progress.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &TaskHeartbeat{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskHeartbeat) DoRun(args []string) (err error) {
	if len(args) < 3 {
		return errShowHelp("Not enough arguments, see below for usage.")
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

	out, err := bclient.SendRunHeartbeat(projectID, args[0], taskID, args[2])
	if err != nil {
		return HandleError(err)
	}

	cmd.outputter.Logger.Printf("Task Heartbeat: %v", out)
	return nil
}
