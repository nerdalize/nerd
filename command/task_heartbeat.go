package command

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

//TaskHeartbeatOpts describes command options
type TaskHeartbeatOpts struct {
	NerdOpts
}

//TaskHeartbeat command
type TaskHeartbeat struct {
	*command
	opts   *TaskHeartbeatOpts
	parser *flags.Parser
}

//TaskHeartbeatFactory returns a factory method for the join command
func TaskHeartbeatFactory() (cli.Command, error) {
	cmd := &TaskHeartbeat{
		command: &command{
			help:     "",
			synopsis: "indicate that a task run is still in progress",
			parser:   flags.NewNamedParser("nerd task heartbeat <queue-id> <task-id> <run-token>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &TaskHeartbeatOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskHeartbeat) DoRun(args []string) (err error) {
	if len(args) < 3 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	config, err := conf.Read()
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	bclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	taskID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		HandleError(errors.Wrap(err, "invalid task ID, must be a number"), cmd.opts.VerboseOutput)
	}

	out, err := bclient.SendRunHeartbeat(config.CurrentProject.Name, args[0], taskID, args[2])
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	logrus.Infof("Task Heartbeat: %v", out)
	return nil
}
