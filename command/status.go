package command

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/client"
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
	c := client.NewNerdAPI(cmd.opts.NerdAPIConfig())
	tasks, aerr := c.ListTasks()
	if aerr != nil {
		err = HandleClientError(aerr, cmd.opts.VerboseOutput)
		return HandleError(err, cmd.opts.VerboseOutput)
	}

	for _, t := range tasks {
		fmt.Printf("%s (%s@%s): %s\n", t.ID, t.Image, t.Dataset, t.Status)
	}

	return nil
}
