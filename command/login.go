package command

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//LoginOpts describes command options
type LoginOpts struct{}

//Login command
type Login struct {
	*command

	ui     cli.Ui
	opts   *LoginOpts
	parser *flags.Parser
}

//LoginFactory returns a factory method for the join command
func LoginFactory() func() (cmd cli.Command, err error) {
	cmd := &Login{
		command: &command{
			help:     "",
			synopsis: "...",
			parser:   flags.NewNamedParser("nerd login", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &LoginOpts{},
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
func (cmd *Login) DoRun(args []string) (err error) {
	return nil
}
