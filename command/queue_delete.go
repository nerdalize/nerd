package command

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
)

//QueueDeleteOpts describes command options
type QueueDeleteOpts struct {
	NerdOpts
}

//QueueDelete command
type QueueDelete struct {
	*command
	opts   *QueueDeleteOpts
	parser *flags.Parser
}

//QueueDeleteFactory returns a factory method for the join command
func QueueDeleteFactory() (cli.Command, error) {
	cmd := &QueueDelete{
		command: &command{
			help:     "",
			synopsis: "...",
			parser:   flags.NewNamedParser("nerd queue delete", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &QueueDeleteOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *QueueDelete) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	config, err := conf.Read()
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	bclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	out, err := bclient.DeleteQueue(config.CurrentProject.Name, args[0])
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	logrus.Infof("Queue Deletion: %v", out)
	return nil
}
