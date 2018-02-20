package cmd

import (
	"context"

	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"

	"github.com/nerdalize/nerd/pkg/transfer"
	"github.com/nerdalize/nerd/svc"

	"github.com/mitchellh/cli"
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

	kube := svc.NewKube(deps)
	mgr, sto, sta, err := cmd.TransferOpts.TransferManager(kube)
	if err != nil {
		return errors.Wrap(err, "failed to setup transfer manager")
	}

	ctx := context.Background()
	var h transfer.Handle
	if h, err = mgr.Create(
		ctx,
		cmd.Name,
		*sto,
		*sta,
	); err != nil {
		return errors.Wrap(err, "failed to create transfer handle")
	}

	defer h.Close()

	err = h.Push(ctx, args[0], transfer.NewDiscardReporter())
	if err != nil {
		return errors.Wrap(err, "failed to upload dataset")
	}

	cmd.out.Infof("Uploaded dataset: '%s'", h.Name())
	cmd.out.Infof("To run a job with a dataset, use: 'nerd job run'")
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
