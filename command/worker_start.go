package command

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
)

//WorkerStartOpts describes command options
type WorkerStartOpts struct {
	NerdOpts
}

//WorkerStart command
type WorkerStart struct {
	*command
	opts   *WorkerStartOpts
	parser *flags.Parser
}

//WorkerStartFactory returns a factory method for the join command
func WorkerStartFactory() (cli.Command, error) {
	cmd := &WorkerStart{
		command: &command{
			help:     "",
			synopsis: "provision a new worker to provide compute",
			parser:   flags.NewNamedParser("nerd worker start", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &WorkerStartOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkerStart) DoRun(args []string) (err error) {
	config, err := conf.Read()
	if err != nil {
		HandleError(err)
	}

	bclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err)
	}

	worker, err := bclient.StartWorker(config.CurrentProject.Name)
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Worker Started: %v", worker)
	return nil
}
