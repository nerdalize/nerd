package cmd

import (
	"context"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

//DatasetDelete command
type DatasetDelete struct {
	KubeOpts

	*command
}

//DatasetDeleteFactory creates the command
func DatasetDeleteFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &DatasetDelete{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, flags.None)
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *DatasetDelete) Execute(args []string) (err error) {
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

	in := &svc.DeleteDatasetInput{
		Name: args[0],
	}

	kube := svc.NewKube(deps)
	_, err = kube.DeleteDataset(ctx, in)
	if err != nil {
		return renderServiceError(err, "failed to delete dataset")
	}

	cmd.out.Infof("Deleted dataset: '%s'", in.Name)
	cmd.out.Infof("To see whats happening, use: 'nerd dataset list'")
	return nil
}

// Description returns long-form help text
func (cmd *DatasetDelete) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *DatasetDelete) Synopsis() string { return "Remove a dataset from the cluster" }

// Usage shows usage
func (cmd *DatasetDelete) Usage() string { return "nerd dataset delete [NAME]" }
