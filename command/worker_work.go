package command

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	nerdaws "github.com/nerdalize/nerd/nerd/aws"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/nerd/service/working/v1"
)

//WorkerWorkOpts describes command options
type WorkerWorkOpts struct {
	NerdOpts
}

//WorkerWork command
type WorkerWork struct {
	*command
	opts   *WorkerWorkOpts
	parser *flags.Parser
}

//WorkerWorkFactory returns a factory method for the join command
func WorkerWorkFactory() (cli.Command, error) {
	cmd := &WorkerWork{
		command: &command{
			help:     "",
			synopsis: "start working tasks of a queue locally",
			parser:   flags.NewNamedParser("nerd worker work <queue-id> <command-tmpl> [arg-tmpl...]", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &WorkerWorkOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkerWork) DoRun(args []string) (err error) {
	if len(args) < 2 {
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

	creds := nerdaws.NewNerdalizeCredentials(bclient, config.CurrentProject.Name)
	qops, err := nerdaws.NewQueueClient(creds, "eu-west-1")
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	logger := log.New(os.Stderr, "worker/", log.Lshortfile)
	conf := v1working.DefaultConf()

	worker := v1working.NewWorker(logger, bclient, qops, config.CurrentProject.Name, args[0], args[1], args[2:], conf)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go worker.Start(ctx)

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)
	<-exitCh

	return nil
}
