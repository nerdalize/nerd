package command

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
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
	config, err := conf.Read()
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	bclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	out, err := bclient.CreateQueue(config.CurrentProject)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	logrus.Infof("Queue Creation: %v", out)
	return nil
}
