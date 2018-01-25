package cmd

import (
	"context"

	"github.com/jessevdk/go-flags"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

//DatasetUpload command
type DatasetUpload struct {
	KubeOpts
	TransferOpts
	Name string `long:"name" short:"n" description:"assign a name to the dataset"`

	*command
}

//DatasetUploadFactory creates the command
func DatasetUploadFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &DatasetUpload{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, flags.PassAfterNonOption)
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *DatasetUpload) Execute(args []string) (err error) {
	if len(args) < 1 {
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

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	ref, err := trans.Upload(ctx, args[0])
	if err != nil {
		return errors.Wrap(err, "failed to perform upload")
	}

	in := &svc.CreateDatasetInput{
		Name:   cmd.Name,
		Bucket: ref.Bucket,
		Key:    ref.Key,
	}

	kube := svc.NewKube(deps)
	out, err := kube.CreateDataset(ctx, in)
	if err != nil {
		return renderServiceError(err, "failed to upload dataset")
	}

	cmd.out.Infof("Upload dataset: '%s'", out.Name)
	cmd.out.Infof("To see available datasets, use: 'nerd dataset list'")
	return nil
}

// Description returns long-form help text
func (cmd *DatasetUpload) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *DatasetUpload) Synopsis() string { return "Upload a dataset to your compute cluster." }

// Usage shows usage
func (cmd *DatasetUpload) Usage() string {
	return "nerd dataset upload [--name=] ~/my-project/my-input-1"
}
