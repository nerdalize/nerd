package command

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
)

//TaskListOpts describes command options
type TaskListOpts struct {
	NerdOpts
}

//TaskList command
type TaskList struct {
	*command
	opts   *TaskListOpts
	parser *flags.Parser
}

//TaskListFactory returns a factory method for the join command
func TaskListFactory() (cli.Command, error) {
	cmd := &TaskList{
		command: &command{
			help:     "",
			synopsis: "...",
			parser:   flags.NewNamedParser("nerd task list <queue-id>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &TaskListOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskList) DoRun(args []string) (err error) {
	if len(args) < 1 {
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

	out, err := bclient.ListTasks(config.CurrentProject.Name, args[0])
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	logrus.Infof("Task Listing: %v", out)
	return nil
}
