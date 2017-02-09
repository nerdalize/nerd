package command

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/client"
)

//LogsOpts describes command options
type LogsOpts struct {
	*NerdAPIOpts
}

//Logs command
type Logs struct {
	*command

	ui     cli.Ui
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
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	c := client.NewNerdAPI(cmd.opts.NerdAPIConfig())

	lines, err := c.ListTaskLogs(args[0])
	if err != nil {
		return fmt.Errorf("failed to list logs: %v", err)
	}

	for _, line := range lines {
		fmt.Println(line)
	}

	return nil
}
