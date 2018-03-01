package cmd

import (
	"context"
	"fmt"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
)

//JobDelete command
type JobDelete struct {
	KubeOpts

	*command
}

//JobDeleteFactory creates the command
func JobDeleteFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &JobDelete{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, flags.None, "nerd job delete")
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *JobDelete) Execute(args []string) (err error) {
	if len(args) < 1 {
		return errShowUsage(fmt.Sprintf(MessageNotEnoughArguments, 1, ""))
	}

	kopts := cmd.KubeOpts
	deps, err := NewDeps(cmd.Logger(), kopts)
	if err != nil {
		return renderConfigError(err, "failed to configure")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	kube := svc.NewKube(deps)
	for i := range args {
		in := &svc.DeleteJobInput{
			Name: args[i],
		}

		_, err = kube.DeleteJob(ctx, in)
		if err != nil {
			return renderServiceError(err, fmt.Sprintf("failed to delete job `%s`", in.Name))
		}

		cmd.out.Infof("Deleted job: '%s'", in.Name)
	}
	cmd.out.Infof("To see whats happening, use: 'nerd job list'")
	return nil
}

// Description returns long-form help text
func (cmd *JobDelete) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *JobDelete) Synopsis() string { return "Remove one or more job(s) from the cluster" }

// Usage shows usage
func (cmd *JobDelete) Usage() string { return "nerd job delete JOB [JOB...]" }
