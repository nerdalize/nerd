package command

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//WorkerStartOpts describes command options
type WorkerStartOpts struct {
	NerdOpts
	Verb func(bool) `short:"a" long:"verb" default:"false" optional:"true" optional-value:"true" description:"show verbose output"`
}

//WorkerStart command
type WorkerStart struct {
	*command
	opts   *WorkerStartOpts
	parser *flags.Parser
}

//WorkerStartFactory returns a factory method for the join command
func WorkerStartFactory() (cli.Command, error) {
	cmd := &WorkerStart{
		command: &command{
			help:     "",
			synopsis: "provision a new worker to provide compute",
			parser:   flags.NewNamedParser("nerd worker start", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &WorkerStartOpts{},
	}

	cmd.opts.Verb = func(set bool) {
		fmt.Println("!!!!!", set)
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkerStart) DoRun(args []string) (err error) {
	return fmt.Errorf("not yet implemented")
}
