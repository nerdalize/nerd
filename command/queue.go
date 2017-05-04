package command

import (
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//Queue command
type Queue struct {
	*command
}

//QueueFactory returns a factory method for the join command
func QueueFactory() (cli.Command, error) {
	cmd := &Queue{
		command: &command{
			help:     `setup queues that transport tasks to workers`,
			synopsis: "setup queues that transport tasks to workers",
			parser:   flags.NewNamedParser("nerd queue", flags.Default),
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
func (cmd *Queue) DoRun(args []string) (err error) {
	return errShowHelp
}
