package command

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//LogsOpts describes command options
type LogsOpts struct {
	*NerdAPIOpts
	*OutputOpts
}

//Logs command
type Logs struct {
	*command

	opts   *LogsOpts
	parser *flags.Parser
}

//LogsFactory returns a factory method for the join command
func LogsFactory() func() (cmd cli.Command, err error) {
	cmd := &Logs{
		command: &command{
			help:     "",
			synopsis: "retrieve up-to-date feedback from a task",
			parser:   flags.NewNamedParser("nerd logs <task_id>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &LogsOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//DoRun is called by run and allows an error to be returned
func (cmd *Logs) DoRun(args []string) (err error) {
	return nil
}
