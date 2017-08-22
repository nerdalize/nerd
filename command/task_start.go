package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//TaskStartOpts describes command options
type TaskStartOpts struct {
	Env []string `long:"env" short:"e" description:"environment variables to"`
}

//TaskStart command
type TaskStart struct {
	*command
	opts *TaskStartOpts
}

//TaskStartFactory returns a factory method for the join command
func TaskStartFactory() (cli.Command, error) {
	opts := &TaskStartOpts{}
	comm, err := newCommand("nerd task start <workload-id> -- [cmd [args...]]", "start a new task for a workload", "", opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &TaskStart{
		command: comm,
		opts:    opts,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskStart) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return errors.Wrap(errShowHelp, "Not enough arguments, see below for usage.")
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	tcmd := []string{}
	if len(args) > 1 {
		tcmd = args[1:]
	}

	tenv := map[string]string{}
	for _, l := range cmd.opts.Env {
		split := strings.SplitN(l, "=", 2)
		if len(split) < 2 {
			return HandleError(fmt.Errorf("invalid environment variable format, expected 'FOO=bar' fromat, got: %v", l))
		}
		tenv[split[0]] = split[1]
	}

	buf := bytes.NewBuffer(nil)
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		lr := io.LimitReader(os.Stdin, 128*1024) //128KiB
		_, err = io.Copy(buf, lr)
		if err != nil {
			return HandleError(fmt.Errorf("failed to copy stdin: %v", err))
		}
	}

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}

	projectID, err := ss.RequireProjectID()
	if err != nil {
		return HandleError(err)
	}

	out, err := bclient.StartTask(projectID, args[0], tcmd, tenv, buf.Bytes())
	if err != nil {
		return HandleError(err)
	}

	cmd.outputter.Logger.Printf("Started task with ID: %d", out.TaskID)
	return nil
}
