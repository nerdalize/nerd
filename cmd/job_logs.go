package cmd

import (
	"bytes"
	"context"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/pkg/kubevisor"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

//JobLogs command
type JobLogs struct {
	KubeOpts
	Tail int64 `long:"tail" short:"t" description:"only return the oldest N lines of the process logs"`

	*command
}

//JobLogsFactory creates the command
func JobLogsFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &JobLogs{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, flags.None)
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *JobLogs) Execute(args []string) (err error) {
	if len(args) < 1 {
		return errShowUsage(MessageNotEnoughArguments)
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
		Tail: cmd.Tail,
	}

	kube := svc.NewKube(deps)
	out, err := kube.FetchJobLogs(ctx, in)
	if err != nil {
		return renderServiceError(err, "failed to fetch job logs")
	}

	lines := string(bytes.TrimSpace(out.Data))
	if len(lines) < 1 {
		cmd.out.Info("-- no visible logs returned --")
		return nil
	}

	cmd.out.Output(string(out.Data))
	if int64(len(out.Data)) == kubevisor.MaxLogBytes {
		cmd.out.Info("-- logs are trimmed after this point --")
	}

	return nil
}

// Description returns long-form help text
func (cmd *JobLogs) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *JobLogs) Synopsis() string { return "Return logs for a running job" }

// Usage shows usage
func (cmd *JobLogs) Usage() string { return "nerd job logs [NAME]" }
