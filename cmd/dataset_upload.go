package cmd

import (
	"context"

	"github.com/jessevdk/go-flags"
	uuid "github.com/satori/go.uuid"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/pkg/transfer"
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

func uploadToDataset(ctx context.Context, trans transfer.Transfer, bucket string, kube *svc.Kube, path, datasetName string) (ref *transfer.Ref, name string, err error) {
	ref = &transfer.Ref{
		Bucket: bucket,
		Key:    uuid.NewV4().String() + ".zip", //@TODO move this to a library
	}

	var n int
	if path != "" { //path is optional
		n, err = trans.Upload(ctx, ref, path)
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to perform upload")
		}
	}

	in := &svc.CreateDatasetInput{
		Name:   datasetName,
		Bucket: ref.Bucket,
		Key:    ref.Key,
		Size:   uint64(n),
	}

	out, err := kube.CreateDataset(ctx, in)
	if err != nil {
		return nil, "", renderServiceError(err, "failed to upload dataset")
	}

	return ref, out.Name, nil
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

	kube := svc.NewKube(deps)

	_, name, err := uploadToDataset(ctx, trans, cmd.AWSS3Bucket, kube, args[0], cmd.Name)
	if err != nil {
		return err
	}

	cmd.out.Infof("Uploaded dataset: '%s'", name)
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
