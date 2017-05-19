package command

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mattn/go-isatty"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
)

//TaskStartOpts describes command options
type TaskStartOpts struct {
	NerdOpts
	Env []string `long:"env" short:"e" description:"environment variables to"`
}

//TaskStart command
type TaskStart struct {
	*command
	opts   *TaskStartOpts
	parser *flags.Parser
}

//TaskStartFactory returns a factory method for the join command
func TaskStartFactory() (cli.Command, error) {
	cmd := &TaskStart{
		command: &command{
			help:     "",
			synopsis: "schedule a new task for workers to consume from a queue",
			parser:   flags.NewNamedParser("nerd task start <queue-id> [<cmd_arg1>, <cmd_arg2>]", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &TaskStartOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskStart) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	config, err := conf.Read()
	if err != nil {
		HandleError(err)
	}

	bclient, err := NewClient(cmd.ui)
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

	out, err := bclient.StartTask(config.CurrentProject.Name, args[0], tcmd, tenv, buf.Bytes())
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Task Start: %v", out)
	return nil
}
