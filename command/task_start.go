package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
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
	comm, err := newCommand("nerd task start <queue-id> [<cmd_arg1>, <cmd_arg2>]", "schedule a new task for workers to consume from a queue", "", opts)
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
		return fmt.Errorf("not enough arguments, see --help")
	}

	bclient, err := NewClient(cmd.ui, cmd.config, cmd.session)
	if err != nil {
		HandleError(err)
	}

	tcmd := []string{}
	if len(args) > 1 {
		tcmd = args[1:]
	}

	tenv := map[string]string{}
	for _, l := range cmd.opts.Env {
		split := strings.SplitN(l, "=", 2)
		if len(split) < 2 {
			HandleError(fmt.Errorf("invalid environment variable format, expected 'FOO=bar' fromat, got: %v", l))
		}
		tenv[split[0]] = split[1]
	}

	buf := bytes.NewBuffer(nil)
	if !isatty.IsTerminal(os.Stdin.Fd()) {
		lr := io.LimitReader(os.Stdin, 128*1024) //128KiB
		_, err = io.Copy(buf, lr)
		if err != nil {
			HandleError(fmt.Errorf("failed to copy stdin: %v", err))
		}
	}

	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}
	out, err := bclient.StartTask(ss.Project.Name, args[0], tcmd, tenv, buf.Bytes())
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Task Start: %v", out)
	return nil
}
