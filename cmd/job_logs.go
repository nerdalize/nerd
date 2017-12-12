package cmd

import (
	"bytes"
	"context"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

//JobLogs command
type JobLogs struct {
	KubeOpts

	*command
}

//JobLogsFactory creates the command
func JobLogsFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &JobLogs{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd)
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *JobLogs) Execute(args []string) (err error) {
	if len(args) < 1 {
		return errors.New(MessageNotEnoughArguments)
	}

	kopts := cmd.KubeOpts
	deps, err := NewDeps(cmd.Logger(), kopts)
	if err != nil {
		return errors.Wrap(err, "failed to configure")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	in := &svc.FetchJobLogsInput{
		Name: args[0],
	}

	kube := svc.NewKube(deps, kopts.Namespace)
	out, err := kube.FetchJobLogs(ctx, in)
	if err != nil {
		return renderServiceError(err, "failed to delete job")
	}

	lines := string(bytes.TrimSpace(out.Data))
	if len(lines) < 1 {
		cmd.out.Errorf("No logs available (anymore) for job '%s'. Maybe the process didn't output any logs or it was created a long time ago: old logs may be discarded", in.Name)
		return nil
	}

	cmd.out.Output(string(out.Data))
	return nil
}

// Description returns long-form help text
func (cmd *JobLogs) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *JobLogs) Synopsis() string { return "Return logs for a running job" }

// Usage shows usage
func (cmd *JobLogs) Usage() string { return "nerd job logs [NAME]" }
