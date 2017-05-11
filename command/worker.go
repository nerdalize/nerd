package command

import (
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//Worker command
type Worker struct {
	*command
}

//WorkerFactory returns a factory method for the join command
func WorkerFactory() (cli.Command, error) {
	cmd := &Worker{
		command: &command{
			help:     `control compute capacity for working on tasks`,
			synopsis: "control compute capacity for working on tasks",
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
func (cmd *Worker) DoRun(args []string) (err error) {
	return errShowHelp
}
