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

//TaskStopOpts describes command options
type TaskStopOpts struct {
	NerdOpts
}

//TaskStop command
type TaskStop struct {
	*command
	opts   *TaskStopOpts
	parser *flags.Parser
}

//TaskStopFactory returns a factory method for the join command
func TaskStopFactory() (cli.Command, error) {
	cmd := &TaskStop{
		command: &command{
			help:     "",
			synopsis: "abort any run(s) of the specified task on a queue",
			parser:   flags.NewNamedParser("nerd task stop <queue-id> <task-id>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &TaskStopOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskStop) DoRun(args []string) (err error) {
	if len(args) < 2 {
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

	out, err := bclient.StopTask(config.CurrentProject.Name, args[0], taskID)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	logrus.Infof("Task Stop: %v", out)
	return nil
}
