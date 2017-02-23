package command

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/client"
	"github.com/nerdalize/nerd/nerd/client/credentials"
)

//StatusOpts describes command options
type StatusOpts struct {
	*NerdAPIOpts
	*OutputOpts
}

//Status command
type Status struct {
	*command

	ui     cli.Ui
	opts   *StatusOpts
	parser *flags.Parser
}

//StatusFactory returns a factory method for the join command
func StatusFactory() func() (cmd cli.Command, err error) {
	cmd := &Status{
		command: &command{
			help:     "",
			synopsis: "show the status of all queued tasks",
			parser:   flags.NewNamedParser("nerd status", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &StatusOpts{},
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
func (cmd *Status) DoRun(args []string) (err error) {
	c, err := client.NewNerdAPI(credentials.NewNerdAPI())
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	tasks, err := c.ListTasks()
	if err != nil {
		return HandleError(HandleClientError(err, cmd.opts.VerboseOutput), cmd.opts.VerboseOutput)
	}

	for _, t := range tasks.Tasks {
		fmt.Printf("%s\n", t.ID)
	}

	return nil
}
