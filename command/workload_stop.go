package command

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//WorkloadStop command
type WorkloadStop struct {
	*command
}

//WorkloadStopFactory returns a factory method for the join command
func WorkloadStopFactory() (cli.Command, error) {
	comm, err := newCommand("nerd workload stop <workload-id>", "stop a workload from providing compute capacity", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &WorkloadStop{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkloadStop) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}

	out, err := bclient.StopWorkload(ss.Project.Name, args[0])
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Workload stopped: %v", out)
	return nil
}
