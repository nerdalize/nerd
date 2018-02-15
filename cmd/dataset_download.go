package cmd

import (
	"context"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	transfer "github.com/nerdalize/nerd/pkg/transfer/v2"
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

	kube := svc.NewKube(deps)
	var mgr transfer.Manager
	if mgr, err = transfer.NewKubeManager(
		kube,
		map[transfer.StoreType]transfer.StoreFactory{
			transfer.StoreTypeS3: transfer.CreateS3Store,
		},
		map[transfer.ArchiverType]transfer.ArchiverFactory{
			transfer.ArchiverTypeTar: transfer.CreateTarArchiver,
		},
	); err != nil {
		return errors.Wrap(err, "failed to setup transfer manager")
	}

	ctx := context.Background()
	var h transfer.Handle
	if h, err = mgr.Open(
		ctx,
		args[0],
	); err != nil {
		return errors.Wrap(err, "failed to create transfer handle")
	}

	defer h.Close()

	err = h.Pull(ctx, args[1], nil)
	if err != nil {
		return errors.Wrap(err, "failed to download dataste")
	}

	cmd.out.Infof("Downloaded dataset: '%s'", h.Name())
	cmd.out.Infof("To delete the dataset from the cloud, use: `nerd dataset delete %s`", h.Name())
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
