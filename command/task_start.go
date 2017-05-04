package command

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

//TaskStartOpts describes command options
type TaskStartOpts struct {
	NerdOpts
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
			parser:   flags.NewNamedParser("nerd task start <queue-id> <payload>", flags.Default),
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

	v := map[string]interface{}{}
	err = json.Unmarshal([]byte(args[1]), &v)
	if err != nil {
		HandleError(errors.Errorf("payload must a valid JSON map, received: '%s'", args[1]), cmd.opts.VerboseOutput)
	}

	out, err := bclient.StartTask(config.CurrentProject, args[0], args[1])
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	logrus.Infof("Task Start: %v", out)
	return nil
}
