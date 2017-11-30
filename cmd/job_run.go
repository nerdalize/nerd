package cmd

import (
	"context"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

//JobRun command
type JobRun struct {
	KubeOpts
	Name string `long:"name" short:"n" description:"assign a name to the job"`

	*command
}

//JobRunFactory creates the command
func JobRunFactory() cli.CommandFactory {
	cmd := &JobRun{}
	cmd.command = createCommand(cmd.Execute, cmd.Description, cmd.Usage, cmd)
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *JobRun) Execute(args []string) (err error) {
	if len(args) < 1 {
		return errors.New(MessageNotEnoughArguments)
	}

	kopts := cmd.KubeOpts
	deps, err := NewDeps(kopts)
	if err != nil {
		return errors.Wrap(err, "failed to configure")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	in := &svc.RunJobInput{
		Image: args[0],
		Name:  cmd.Name,
	}

	kube := svc.NewKube(deps, kopts.Namespace)
	out, err := kube.RunJob(ctx, in)
	if err != nil {
		return errors.Wrap(err, "failed to run job")
	}

	//@TODO find a way of formatting the output

	cmd.logs.Printf("%#v", out)
	return nil
}

// Description returns long-form help text
func (cmd *JobRun) Description() string { return PlaceholderHelp }

// Synopsis returns a one-line
func (cmd *JobRun) Synopsis() string { return PlaceholderSynopsis }

// Usage shows usage
func (cmd *JobRun) Usage() string { return PlaceholderUsage }
