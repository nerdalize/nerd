package command

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//WorkerLogs command
type WorkerLogs struct {
	*command
}

//WorkerLogsFactory returns a factory method for the join command
func WorkerLogsFactory() (cli.Command, error) {
	comm, err := newCommand("nerd worker logs <workload-id> <worker-id>", "Return recent logs from a worker.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &WorkerLogs{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkerLogs) DoRun(args []string) (err error) {
	if len(args) < 2 {
		return errShowHelp("Not enough arguments, see below for usage.")
	}

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

	out, err := bclient.WorkerLogs(projectID, args[0], args[1])
	if err != nil {
		return HandleError(err)
	}

	fmt.Fprintf(os.Stdout, "%s", string(out.Data))
	return nil
}
