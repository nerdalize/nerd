package command

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
	// <<<<<<< HEAD
	// 	"github.com/Sirupsen/logrus"
	// 	"github.com/jessevdk/go-flags"
	// 	"github.com/mitchellh/cli"
	// 	"github.com/nerdalize/nerd/nerd/conf"
	// =======
	// 	"github.com/mitchellh/cli"
	// 	"github.com/pkg/errors"
	// >>>>>>> master
)

//WorkerStop command
type WorkerStop struct {
	*command
}

//WorkerStopFactory returns a factory method for the join command
func WorkerStopFactory() (cli.Command, error) {
	comm, err := newCommand("nerd worker stop <worker_id>", "stop a worker from providing compute capacity", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &WorkerStart{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkerStop) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	bclient, err := NewClient(cmd.ui, cmd.config, cmd.session)
	if err != nil {
		HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}

	out, err := bclient.StopWorker(ss.Project.Name, args[0])
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Worker stopped: %v", out)
	return nil
}
