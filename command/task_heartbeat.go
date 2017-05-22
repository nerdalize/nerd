package command

import (
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//TaskHeartbeat command
type TaskHeartbeat struct {
	*command
}

//TaskHeartbeatFactory returns a factory method for the join command
func TaskHeartbeatFactory() (cli.Command, error) {
	comm, err := newCommand("nerd task heartbeat <queue-id> <task-id> <run-token>", "indicate that a task run is still in progress", "", nil)
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
		return fmt.Errorf("not enough arguments, see --help")
	}

	bclient, err := NewClient(cmd.ui, cmd.config, cmd.session)
	if err != nil {
		HandleError(err)
	}

	taskID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		HandleError(errors.Wrap(err, "invalid task ID, must be a number"))
	}

	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}
	out, err := bclient.SendRunHeartbeat(ss.Project.Name, args[0], taskID, args[2])
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Task Heartbeat: %v", out)
	return nil
}
