package cmd

import (
	"context"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

//JobDelete command
type JobDelete struct {
	KubeOpts

	*command
}

//JobDeleteFactory creates the command
func JobDeleteFactory() cli.CommandFactory {
	cmd := &JobDelete{}
	cmd.command = createCommand(cmd.Execute, cmd.Description, cmd.Usage, cmd)
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *JobDelete) Execute(args []string) (err error) {
	if len(args) < 1 {
		return errors.New(MessageNotEnoughArguments)
	}

	kopts := cmd.KubeOpts
	deps, err := NewDeps(cmd.logs, kopts)
	if err != nil {
		return errors.Wrap(err, "failed to configure")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	in := &svc.DeleteJobInput{
		Name: args[0],
	}

	kube := svc.NewKube(deps)
	_, err = kube.DeleteJob(ctx, in)
	if err != nil {
		return errors.Wrap(err, "failed to run job")
	}

	cmd.logs.Printf("Deleted job: '%s'", in.Name)
	return nil
}

// Description returns long-form help text
func (cmd *JobDelete) Description() string { return PlaceholderHelp }

// Synopsis returns a one-line
func (cmd *JobDelete) Synopsis() string { return PlaceholderSynopsis }

// Usage shows usage
func (cmd *JobDelete) Usage() string { return PlaceholderUsage }
