package cmd

import (
	"context"
	"fmt"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
)

//DatasetDelete command
type DatasetDelete struct {
	KubeOpts

	*command
}

//DatasetDeleteFactory creates the command
func DatasetDeleteFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &DatasetDelete{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, flags.None, "nerd dataset delete")
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *DatasetDelete) Execute(args []string) (err error) {
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
		in := &svc.DeleteDatasetInput{
			Name: args[i],
		}

		_, err = kube.DeleteDataset(ctx, in)
		if err != nil {
			return renderServiceError(err, fmt.Sprintf("failed to delete dataset `%s`", in.Name))
		}

		cmd.out.Infof("Deleted dataset: '%s'", in.Name)
	}
	return nil
}

// Description returns long-form help text
func (cmd *DatasetDelete) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *DatasetDelete) Synopsis() string { return "Remove a dataset from the cluster" }

// Usage shows usage
func (cmd *DatasetDelete) Usage() string { return "nerd dataset delete DATASET_NAME [DATASET_NAME...]" }
