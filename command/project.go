package command

import (
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//Project command
type Project struct {
	*command
}

//ProjectFactory returns a factory method for the join command
func ProjectFactory() (cli.Command, error) {
	cmd := &Queue{
		command: &command{
			help:     "",
			synopsis: "set and list projects",
			parser:   flags.NewNamedParser("nerd project <subcommand>", flags.Default),
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
func (cmd *Project) DoRun(args []string) (err error) {
	return errShowHelp
}
