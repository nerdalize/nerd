package command

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//QueueCreateOpts describes command options
type QueueCreateOpts struct {
	NerdOpts
}

//QueueCreate command
type QueueCreate struct {
	*command
	opts   *QueueCreateOpts
	parser *flags.Parser
}

//QueueCreateFactory returns a factory method for the join command
func QueueCreateFactory() (cli.Command, error) {
	cmd := &QueueCreate{
		command: &command{
			help:     "",
			synopsis: "...",
			parser:   flags.NewNamedParser("nerd queue create", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &QueueCreateOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *QueueCreate) DoRun(args []string) (err error) {

	return nil
}
