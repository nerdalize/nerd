package cmd

import (
	"context"
	"fmt"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
)

//DatasetDelete command
type DatasetDelete struct {
	KubeOpts
	All bool `long:"all" short:"a" description:"delete all your datasets in one command"`

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
	if cmd.All {
		return cmd.deleteAll()
	}
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

func (cmd *DatasetDelete) deleteAll() error {
	kopts := cmd.KubeOpts
	deps, err := NewDeps(cmd.Logger(), kopts)
	if err != nil {
		return renderConfigError(err, "failed to configure")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	s, err := cmd.out.Ask("Are you sure you want to delete all your datasets? (y/N)")
	if err != nil {
		return err
	}
	if !strings.HasPrefix(strings.ToLower(s), "y") {
		return nil
	}

	kube := svc.NewKube(deps)
	datasets, err := kube.ListDatasets(ctx, &svc.ListDatasetsInput{})
	if err != nil {
		return renderServiceError(err, "failed to get all datasets")
	}
	if len(datasets.Items) == 0 {
		cmd.out.Info("No dataset found.")
	}
	for _, ds := range datasets.Items {
		in := &svc.DeleteDatasetInput{
			Name: ds.Name,
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
