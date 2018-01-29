package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/jessevdk/go-flags"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
)

//JobRun command
type JobRun struct {
	KubeOpts
	Name string   `long:"name" short:"n" description:"assign a name to the job"`
	Env  []string `long:"env" short:"e" description:"environment variables to use"`

	*command
}

//JobRunFactory creates the command
func JobRunFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &JobRun{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, flags.PassAfterNonOption)
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *JobRun) Execute(args []string) (err error) {
	if len(args) < 1 {
		return errShowUsage(MessageNotEnoughArguments)
	}

	kopts := cmd.KubeOpts
	deps, err := NewDeps(cmd.Logger(), kopts)
	if err != nil {
		return renderConfigError(err, "failed to configure")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	jargs := []string{}
	if len(args) > 1 {
		jargs = args[1:]
	}

	jenv := map[string]string{}
	for _, l := range cmd.Env {
		split := strings.SplitN(l, "=", 2)
		if len(split) < 2 {
			return fmt.Errorf("invalid environment variable format, expected 'FOO=bar' format, got: %v", l)
		}
		jenv[split[0]] = split[1]
	}

	in := &svc.RunJobInput{
		Image: args[0],
		Name:  cmd.Name,
		Env:   jenv,
		Args:  jargs,
	}

	kube := svc.NewKube(deps)
	out, err := kube.RunJob(ctx, in)
	if err != nil {
		return renderServiceError(err, "failed to run job")
	}

	cmd.out.Infof("Submitted job: '%s'", out.Name)
	cmd.out.Infof("To see whats happening, use: 'nerd job list'")
	return nil
}

// Description returns long-form help text
func (cmd *JobRun) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *JobRun) Synopsis() string { return "Runs a job on your compute cluster" }

// Usage shows usage
func (cmd *JobRun) Usage() string { return "nerd job run [OPTIONS] IMAGE [COMMAND] [ARG...]" }
