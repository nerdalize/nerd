package command

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//DownloadOpts describes command options
type DownloadOpts struct{}

//Download command
type Download struct {
	*command

	ui     cli.Ui
	opts   *DownloadOpts
	parser *flags.Parser
}

//DownloadFactory returns a factory method for the join command
func DownloadFactory() func() (cmd cli.Command, err error) {
	cmd := &Download{
		command: &command{
			help:     "",
			synopsis: "...",
			parser:   flags.NewNamedParser("nerd upload", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &DownloadOpts{},
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
func (cmd *Download) DoRun(args []string) (err error) {
	return nil
}
