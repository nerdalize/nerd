package command

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/data"
)

//UploadOpts describes command options
type UploadOpts struct {
	*NerdAPIOpts
	*AuthAPIOpts
}

//Upload command
type Upload struct {
	*command

	ui     cli.Ui
	opts   *UploadOpts
	parser *flags.Parser
}

//UploadFactory returns a factory method for the join command
func UploadFactory() func() (cmd cli.Command, err error) {
	cmd := &Upload{
		command: &command{
			help:     "",
			synopsis: "push task data as input to cloud storage",
			parser:   flags.NewNamedParser("nerd upload <dataset> <path>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &UploadOpts{},
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
func (cmd *Upload) DoRun(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	dataset := args[0]
	path := args[1]

	nerdclient := NewClient(cmd.ui, cmd.opts.URL(), cmd.opts.AuthAPIURL)
	client, err := data.NewClient(data.NewNerdalizeCredentials(nerdclient))
	if err != nil {
		return fmt.Errorf("could not create data client: %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("argument '%v' is not a valid file or directory", path)
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		return client.UploadDir(path, dataset, &stdoutkw{}, 64)
	case mode.IsRegular():
		return client.UploadFile(path, path, dataset)
	}
	return nil
}
