package cmd

import (
	"context"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

//JobRun command
type JobRun struct {
	*command
}

//JobRunFactory creates the command
func JobRunFactory() cli.CommandFactory {
	cmd := &JobRun{}
	cmd.command = createCommand(cmd.Execute, cmd.Description, cmd.Usage)
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *JobRun) Execute(args []string) (err error) {
	if len(args) < 1 {
		return errors.New(MessageNotEnoughArguments)
	}

	kube, err := svc.NewKube(args[0])
	if err != nil {
		return errors.Wrap(err, "failed to setup Kubernetes connection")
	}

	ctx := context.Background()
	in := &svc.RunJobInput{}
	out, err := kube.RunJob(ctx, in)
	if err != nil {
		return errors.Wrap(err, "failed to run job")
	}

	cmd.logs.Printf("%#v", out)
	return nil
}

// Description returns long-form help text
func (cmd *JobRun) Description() string { return PlaceholderHelp }

// Synopsis returns a one-line
func (cmd *JobRun) Synopsis() string { return PlaceholderSynopsis }

// Usage shows usage
func (cmd *JobRun) Usage() string { return PlaceholderUsage }
