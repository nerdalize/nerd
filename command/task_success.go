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

//TaskSuccessOpts describes command options
type TaskSuccessOpts struct {
	NerdOpts
}

//TaskSuccess command
type TaskSuccess struct {
	*command
	opts   *TaskSuccessOpts
	parser *flags.Parser
}

//TaskSuccessFactory returns a factory method for the join command
func TaskSuccessFactory() (cli.Command, error) {
	cmd := &TaskSuccess{
		command: &command{
			help:     "",
			synopsis: "...",
			parser:   flags.NewNamedParser("nerd task success <queue-id> <task-id> <run-token> <result>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &TaskSuccessOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskSuccess) DoRun(args []string) (err error) {
	if len(args) < 4 {
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

	out, err := bclient.SendRunSuccess(config.CurrentProject, args[0], taskID, args[2], args[3])
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	logrus.Infof("Task Success: %v", out)
	return nil
}
