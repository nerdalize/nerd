package command

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//WorkloadDescribe command
type WorkloadDescribe struct {
	*command
}

//WorkloadDescribeFactory returns a factory method for the join command
func WorkloadDescribeFactory() (cli.Command, error) {
	comm, err := newCommand("nerd workload describe <workload-id>", "return more information about a specific workload", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &WorkloadDescribe{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkloadDescribe) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}
	out, err := bclient.DescribeWorkload(ss.Project.Name, args[0])
	if err != nil {
		return HandleError(err)
	}

	logrus.Infof("Workload Description: %+v", out)
	return nil
}
