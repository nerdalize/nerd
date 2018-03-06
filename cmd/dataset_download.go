package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	homedir "github.com/mitchellh/go-homedir"
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

	Input  string `long:"input-of" description:"specify a job name where the datasets were used as its input. Dataset name is no longer mandatory."`
	Output string `long:"output-of" description:"specify a job name where the datasets were used as its output. Dataset name is no longer mandatory."`

	*command
}

//DatasetDownloadFactory creates the command
func DatasetDownloadFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &DatasetDownload{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, flags.None, "nerd dataset download")
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *DatasetDownload) Execute(args []string) (err error) {
	var (
		datasetName, outputDir string
	)

	switch l := len(args); {
	case l > 2:
		return errShowUsage(fmt.Sprintf(MessageTooManyArguments, 2, "s"))
	case l == 2:
		datasetName = args[0]
		outputDir = args[1]
	case l == 1:
		if cmd.Input == "" && cmd.Output == "" {
			return errShowUsage(fmt.Sprintf(MessageNotEnoughArguments, 2, "s"))
		}
		outputDir = args[0]
	default:
		return errShowUsage(fmt.Sprintf(MessageNotEnoughArguments, 1, ""))
	}

	//Expand tilde for homedir
	outputDir, err = homedir.Expand(outputDir)
	if err != nil {
		return errors.Wrap(err, "failed to expand home directory in dataset local path")
	}

	outputDir, err = filepath.Abs(outputDir)
	if err != nil {
		return renderServiceError(err, "failed to turn local path into absolute path")
	}

	deps, err := NewDeps(cmd.Logger(), cmd.KubeOpts)
	if err != nil {
		return renderConfigError(err, "failed to configure")
	}

	kube := svc.NewKube(deps)
	var mgr transfer.Manager
	if mgr, err = transfer.NewKubeManager(
		kube,
	); err != nil {
		return errors.Wrap(err, "failed to setup transfer manager")
	}
	ctx := context.Background()

	// if there is only one dataset to download
	if datasetName != "" {
		var h transfer.Handle
		if h, err = mgr.Open(
			ctx,
			datasetName,
		); err != nil {
			return renderServiceError(err, "failed to open dataset '%s'", datasetName)
		}

		defer h.Close()

		err = h.Pull(ctx, outputDir, &progressBarReporter{})
		if err != nil {
			return renderServiceError(err, "failed to download dataset")
		}

		cmd.out.Infof("Downloaded dataset: '%s'", h.Name())
		cmd.out.Infof("To delete the dataset from the cloud, use: `nerd dataset delete %s`", h.Name())
		return nil
	}

	ds, err := kube.ListDatasets(ctx, &svc.ListDatasetsInput{})
	if err != nil {
		return errors.Wrap(err, "failed to download datasets")
	}

	_, err = os.Open(outputDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to open directory")
		}

		err = os.Mkdir(outputDir, 0777) //@TODO decide on permissions before umask
		if err != nil {
			return errors.Wrap(err, "failed to create directory")
		}

		_, err = os.Open(outputDir)
		if err != nil {
			return errors.Wrap(err, "failed open created directory")
		}
	}
	datasets := extractDatasets(ds.Items, cmd.Input, cmd.Output)
	for _, dataset := range datasets {
		dir := outputDir
		if len(datasets) > 1 {
			dir = filepath.Join(outputDir, dataset.Name)
		}
		var h transfer.Handle
		if h, err = mgr.Open(
			ctx,
			dataset.Name,
		); err != nil {
			return errors.Wrap(err, "failed to create transfer handle")
		}

		defer h.Close()

		err = h.Pull(ctx, dir, &progressBarReporter{})
		if err != nil {
			return errors.Wrap(err, "failed to download dataset")
		}
	}
	if len(datasets) == 0 {
		cmd.out.Infof("No dataset found, please check with `nerd job list` or `nerd dataset list` the job name(s) provided.")
	} else if len(datasets) > 1 {
		cmd.out.Infof("Downloaded %d datasets in %s", len(datasets), outputDir)
	} else {
		cmd.out.Infof("Downloaded %d dataset in %s", len(datasets), outputDir)
	}
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
	return "nerd dataset [OPTIONS] download DATASET_NAME DOWNLOAD_PATH"
}

func extractDatasets(ds []*svc.ListDatasetItem, input, output string) map[string]*svc.ListDatasetItem {
	datasets := make(map[string]*svc.ListDatasetItem)

	for _, d := range ds {
		for _, i := range d.Details.InputFor {
			if i == input {
				datasets[d.Name] = d
			}
		}
		for _, o := range d.Details.OutputFrom {
			if o == output {
				datasets[d.Name] = d
			}
		}
	}

	return datasets
}
