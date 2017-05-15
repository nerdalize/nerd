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

//TaskFailureOpts describes command options
type TaskFailureOpts struct {
	NerdOpts
}

//TaskFailure command
type TaskFailure struct {
	*command
	opts   *TaskFailureOpts
	parser *flags.Parser
}

//TaskFailureFactory returns a factory method for the join command
func TaskFailureFactory() (cli.Command, error) {
	cmd := &TaskFailure{
		command: &command{
			help:     "",
			synopsis: "mark a task run as being failed",
			parser:   flags.NewNamedParser("nerd task failure <queue-id> <task-id> <run-token> <error-code> <err-message>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &TaskFailureOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskFailure) DoRun(args []string) (err error) {
	if len(args) < 5 {
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

	taskID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		HandleError(errors.Wrap(err, "invalid task ID, must be a number"))
	}

	out, err := bclient.SendRunFailure(config.CurrentProject.Name, args[0], taskID, args[2], args[3], args[4])
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Task Failure: %v", out)
	return nil
}
