package command

import (
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//TaskSuccess command
type TaskSuccess struct {
	*command
}

//TaskSuccessFactory returns a factory method for the join command
func TaskSuccessFactory() (cli.Command, error) {
	comm, err := newCommand("nerd task success <queue-id> <task-id> <run-token> <result>", "mark a task run as having succeeded", "", nil)
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
	out, err := bclient.SendRunSuccess(ss.Project.Name, args[0], taskID, args[2], args[3])
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Task Success: %v", out)
	return nil
}
