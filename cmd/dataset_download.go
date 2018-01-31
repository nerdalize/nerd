package cmd

import (
	"context"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/pkg/transfer"
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
	TransferOpts

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

	deps, err := NewDeps(cmd.Logger(), cmd.KubeOpts)
	if err != nil {
		return renderConfigError(err, "failed to configure")
	}

	trans, err := cmd.TransferOpts.Transfer()
	if err != nil {
		return errors.Wrap(err, "failed configure transfer")
	}

	//get the dataset by name

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	in := &svc.GetDatasetInput{
		Name: args[0],
	}

	kube := svc.NewKube(deps)
	out, err := kube.GetDataset(ctx, in)
	if err != nil {
		return renderServiceError(err, "failed to download dataset")
	}

	//Use dataset spec to downloa

	ref := &transfer.Ref{
		Bucket: out.Bucket,
		Key:    out.Key,
	}

	err = trans.Download(ctx, ref, args[1])
	if err != nil {
		return errors.Wrap(err, "failed to download")
	}

	cmd.out.Infof("Downloaded dataset: '%s'", out.Name)
	cmd.out.Infof("To delete the dataset from the cloud, use: `nerd dataset delete %s`", out.Name)
	return nil
}

// Description returns long-form help text
func (cmd *DatasetDownload) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *DatasetDownload) Synopsis() string {
	return "Download results from a running job"
}

// Usage shows usage
func (cmd *DatasetDownload) Usage() string {
	return "nerd dataset download <DATASET-NAME> <DOWNLOAD-PATH>"
}
