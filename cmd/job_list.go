package cmd

import (
	"context"
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

//JobList command
type JobList struct {
	KubeOpts
	Name string `long:"name" short:"n" description:"assign a name to the job"`

	*command
}

//JobListFactory creates the command
func JobListFactory() cli.CommandFactory {
	cmd := &JobList{}
	cmd.command = createCommand(cmd.Execute, cmd.Description, cmd.Usage, cmd)
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *JobList) Execute(args []string) (err error) {
	kopts := cmd.KubeOpts
	deps, err := NewDeps(cmd.logs, kopts)
	if err != nil {
		return errors.Wrap(err, "failed to configure")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	in := &svc.ListJobsInput{}
	kube := svc.NewKube(deps, kopts.Namespace)
	out, err := kube.ListJobs(ctx, in)
	if err != nil {
		return errors.Wrap(err, "failed to run job")
	}

	for _, item := range out.Items {
		fmt.Printf("%#v\n", item) //@TODO add proper output formatting
	}

	return nil
}

// Description returns long-form help text
func (cmd *JobList) Description() string { return PlaceholderHelp }

// Synopsis returns a one-line
func (cmd *JobList) Synopsis() string { return PlaceholderSynopsis }

// Usage shows usage
func (cmd *JobList) Usage() string { return PlaceholderUsage }
