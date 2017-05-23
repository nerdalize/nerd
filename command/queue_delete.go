package command

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//QueueDelete command
type QueueDelete struct {
	*command
}

//QueueDeleteFactory returns a factory method for the join command
func QueueDeleteFactory() (cli.Command, error) {
	comm, err := newCommand("nerd queue delete", "remove a queue and all tasks currently in it", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &QueueDelete{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *QueueDelete) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	bclient, err := NewClient(cmd.ui, cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}
	out, err := bclient.DeleteQueue(ss.Project.Name, args[0])
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Queue Deletion: %v", out)
	return nil
}
