package command

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
)

//WorkerStopOpts describes command options
type WorkerStopOpts struct {
	NerdOpts
}

//WorkerStop command
type WorkerStop struct {
	*command
	opts   *WorkerStopOpts
	parser *flags.Parser
}

//WorkerStopFactory returns a factory method for the join command
func WorkerStopFactory() (cli.Command, error) {
	cmd := &WorkerStop{
		command: &command{
			help:     "",
			synopsis: "stop a worker from providing compute capacity",
			parser:   flags.NewNamedParser("nerd worker stop <worker_id>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &WorkerStopOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkerStop) DoRun(args []string) (err error) {
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

	worker, err := bclient.StopWorker(config.CurrentProject.Name, args[0])
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Worker stopped: %v", worker)
	return nil
}
