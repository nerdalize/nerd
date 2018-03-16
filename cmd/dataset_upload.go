package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"

	"github.com/nerdalize/nerd/pkg/transfer"
	"github.com/nerdalize/nerd/svc"

	"github.com/mitchellh/cli"
	homedir "github.com/mitchellh/go-homedir"
)

//DatasetUpload command
type DatasetUpload struct {
	Name string `long:"name" short:"n" description:"assign a name to the dataset"`

	*command
}

//DatasetUploadFactory creates the command
func DatasetUploadFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &DatasetUpload{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, &TransferOpts{}, flags.PassAfterNonOption, "nerd dataset upload")
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *DatasetUpload) Execute(args []string) (err error) {
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	if len(args) < 1 {
		return errShowUsage(fmt.Sprintf(MessageNotEnoughArguments, 1, ""))
	} else if len(args) > 1 {
		return errShowUsage(fmt.Sprintf(MessageTooManyArguments, 1, ""))
	}

	//Expand tilde for homedir
	dir, err := homedir.Expand(args[0])
	if err != nil {
		return renderServiceError(err, "failed to expand home directory in dataset local path")
	}

	dir, err = filepath.Abs(args[0])
	if err != nil {
		return renderServiceError(err, "failed to turn local path into absolute path")
	}

	// check if directory exists
	_, err = os.Open(dir)
	if err != nil {
		return errors.Wrap(err, "failed to upload dataset")
	}

	deps, err := NewDeps(cmd.Logger(), cmd.globalOpts.KubeOpts)
	if err != nil {
		return renderConfigError(err, "failed to configure")
	}

	kube := svc.NewKube(deps)
	t, ok := cmd.advancedOpts.(*TransferOpts)
	if !ok {
		return renderConfigError(fmt.Errorf("unable to use transfer options"), "failed to configure")
	}
	mgr, sto, sta, err := t.TransferManager(kube)
	if err != nil {
		return errors.Wrap(err, "failed to setup transfer manager")
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var h transfer.Handle
	if h, err = mgr.Create(
		ctx,
		cmd.Name,
		*sto,
		*sta,
	); err != nil {
		return renderServiceError(err, "failed to create dataset with name '%s'", cmd.Name)
	}

	defer h.Close()
	go func() {
		<-sigCh
		cancel()
	}()

	err = h.Push(ctx, dir, &progressBarReporter{})
	if err != nil {
		ctx := context.Background() //new context for deletion
		e := mgr.Remove(ctx, h.Name())
		if e != nil {
			return errors.Wrapf(err, "failed to remove dataset: %v", e)
		}

		return renderServiceError(err, "failed to upload dataset")
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
	return "nerd dataset upload [OPTIONS] DIR_TO_UPLOAD"
}
