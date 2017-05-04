package command

import (
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//Task command
type Task struct {
	*command
}

//TaskFactory returns a factory method for the join command
func TaskFactory() (cli.Command, error) {
	cmd := &Task{
		command: &command{
			help:     `manage the lifecycle of compute tasks`,
			synopsis: "manage the lifecycle of compute tasks",
			parser:   flags.NewNamedParser("nerd task", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},
	}

	cmd.runFunc = cmd.DoRun
	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Task) DoRun(args []string) (err error) {
	return errShowHelp
}
