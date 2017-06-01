package command

import (
	"fmt"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//TaskFailure command
type TaskFailure struct {
	*command
}

//TaskFailureFactory returns a factory method for the join command
func TaskFailureFactory() (cli.Command, error) {
	comm, err := newCommand("nerd task failure <workload-id> <task-id> <run-token> <error-code> <err-message>", "mark a task run as being failed", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &TaskFailure{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskFailure) DoRun(args []string) (err error) {
	if len(args) < 5 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
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
	out, err := bclient.SendRunFailure(ss.Project.Name, args[0], taskID, args[2], args[3], args[4])
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Task Failure: %v", out)
	return nil
}
