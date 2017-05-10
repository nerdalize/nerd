package command

import (
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	nerdaws "github.com/nerdalize/nerd/nerd/aws"
	"github.com/nerdalize/nerd/nerd/conf"
)

//TaskReceiveOpts describes command options
type TaskReceiveOpts struct {
	NerdOpts
}

//TaskReceive command
type TaskReceive struct {
	*command
	opts   *TaskReceiveOpts
	parser *flags.Parser
}

//TaskReceiveFactory returns a factory method for the join command
func TaskReceiveFactory() (cli.Command, error) {
	cmd := &TaskReceive{
		command: &command{
			help:     "",
			synopsis: "wait for a new task run to be available on a queue",
			parser:   flags.NewNamedParser("nerd task receive <queue-id>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &TaskReceiveOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskReceive) DoRun(args []string) (err error) {
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

	creds := nerdaws.NewNerdalizeCredentials(bclient, config.CurrentProject.Name)
	qops, err := nerdaws.NewQueueClient(creds, config.CurrentProject.AWSRegion) //@TODO get region from credentials provider
	if err != nil {
		HandleError(err)
	}

	out, err := bclient.ReceiveTaskRuns(config.CurrentProject.Name, args[0], time.Minute*3, qops)
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Task Receiving: %v", out)
	return nil
}
