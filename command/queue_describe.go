package command

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
)

//QueueDescribeOpts describes command options
type QueueDescribeOpts struct {
	NerdOpts
}

//QueueDescribe command
type QueueDescribe struct {
	*command
	opts   *QueueDescribeOpts
	parser *flags.Parser
}

//QueueDescribeFactory returns a factory method for the join command
func QueueDescribeFactory() (cli.Command, error) {
	cmd := &QueueDescribe{
		command: &command{
			help:     "",
			synopsis: "return more information about a specific queue",
			parser:   flags.NewNamedParser("nerd queue describe <queue-id>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &QueueDescribeOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *QueueDescribe) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	config, err := conf.Read()
	if err != nil {
		HandleError(err)
	}

	bclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err)
	}

	out, err := bclient.DescribeQueue(config.CurrentProject.Name, args[0])
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Queue Description: %+v", out)
	return nil
}
