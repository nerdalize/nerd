package command

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/client"
	"github.com/nerdalize/nerd/nerd/client/credentials"
)

//RunOpts describes command options
type RunOpts struct {
	*NerdAPIOpts
	*OutputOpts
}

//Run command
type Run struct {
	*command

	ui     cli.Ui
	opts   *RunOpts
	parser *flags.Parser
}

//RunFactory returns a factory method for the join command
func RunFactory() func() (cmd cli.Command, err error) {
	cmd := &Run{
		command: &command{
			help:     "",
			synopsis: "create a new compute task for a dataset",
			parser:   flags.NewNamedParser("nerd run <image> <dataset>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &RunOpts{},
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
func (cmd *Run) DoRun(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	c, err := client.NewNerdAPI(credentials.NewNerdAPI())
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	_, err = c.CreateTask(args[0], args[1], args[2:])
	if err != nil {
		return HandleError(HandleClientError(err, cmd.opts.VerboseOutput), cmd.opts.VerboseOutput)
	}
	return nil
}
