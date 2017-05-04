package command

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//WorkerWorkOpts describes command options
type WorkerWorkOpts struct {
	NerdOpts
}

//WorkerWork command
type WorkerWork struct {
	*command
	opts   *WorkerWorkOpts
	parser *flags.Parser
}

//WorkerWorkFactory returns a factory method for the join command
func WorkerWorkFactory() (cli.Command, error) {
	cmd := &WorkerWork{
		command: &command{
			help:     "",
			synopsis: "start working tasks on a queue",
			parser:   flags.NewNamedParser("nerd worker work <queue-id>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &WorkerWorkOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkerWork) DoRun(args []string) (err error) {
	return fmt.Errorf("not yet implemented")
}
