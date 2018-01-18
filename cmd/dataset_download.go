package cmd

import (
	"context"
	"fmt"
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

const (
	//OutputDirPermissions are the output directory's permissions.
	OutputDirPermissions = 0755
)

//DatasetDownload command
type DatasetDownload struct {
	KubeOpts
	JobOutput string `long:"job-output" description:"output of the precised job"`
	JobInput  string `long:"job-input" description:"input of the precised job"`

	*command
}

//DatasetDownloadFactory creates the command
func DatasetDownloadFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &DatasetDownload{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, flags.None)
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *DatasetDownload) Execute(args []string) (err error) {
	if len(args) < 2 {
		return errShowUsage(MessageNotEnoughArguments)
	}

	outputDir := args[1]
	// Folder create and check
	fi, err := os.Stat(outputDir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, OutputDirPermissions)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("The provided path '%s' does not exist and could not be created.", outputDir))
		}
		fi, err = os.Stat(outputDir)
	}
	if err != nil {
		return errors.Wrap(err, "The folder could not be created")
	} else if !fi.IsDir() {
		return errors.Errorf("The provided path '%s' is not a directory", outputDir)
	}

	kopts := cmd.KubeOpts
	deps, err := NewDeps(cmd.Logger(), kopts)
	if err != nil {
		return errors.Wrap(err, "failed to configure")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	in := &svc.DownloadDatasetInput{
		JobInput:  cmd.JobInput,
		JobOutput: cmd.JobOutput,
		Name:      args[0],
		Dest:      outputDir,
	}
	kube := svc.NewKube(deps)
	out, err := kube.DownloadDataset(ctx, in)
	if err != nil {
		return renderServiceError(err, "failed to download dataset")
	}

	cmd.out.Infof("Downloaded dataset: '%s'", out.Name)
	cmd.out.Infof("To delete the dataset from the cloud, use: `nerd dataset delete %s`", out.Name)
	return nil
}

// Description returns long-form help text
func (cmd *DatasetDownload) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *DatasetDownload) Synopsis() string {
	return "Return datasets that are managed by the cluster"
}

// Usage shows usage
func (cmd *DatasetDownload) Usage() string { return "nerd dataset list" }
