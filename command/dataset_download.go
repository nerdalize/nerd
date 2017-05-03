package command

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dchest/safefile"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd"
	"github.com/nerdalize/nerd/nerd/aws"
	v1data "github.com/nerdalize/nerd/nerd/client/data/v1"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

const (
	//OutputDirPermissions are the output directory's permissions.
	OutputDirPermissions = 0755
	//DownloadConcurrency is the amount of concurrent download threads.
	DownloadConcurrency = 64
)

//DownloadOpts describes command options
type DownloadOpts struct {
	NerdOpts
	AlwaysOverwrite bool `long:"always-overwrite" default-mask:"false" description:"always overwrite files when they already exist"`
}

//Download command
type Download struct {
	*command

	opts   *DownloadOpts
	parser *flags.Parser
}

//DatasetDownloadFactory returns a factory method for the join command
func DatasetDownloadFactory() (cli.Command, error) {
	cmd := &Download{
		command: &command{
			help:     "",
			synopsis: "Download a dataset from cloud storage",
			parser:   flags.NewNamedParser("nerd dataset download <dataset> <output-dir>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &DownloadOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Download) DoRun(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	config, err := conf.Read()
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	dataset := args[0]
	outputDir := args[1]

	// Folder create and check
	fi, err := os.Stat(outputDir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, OutputDirPermissions)
		if err != nil {
			HandleError(errors.Errorf("The provided path '%s' does not exist and could not be created.", outputDir), cmd.opts.VerboseOutput)
		}
		fi, err = os.Stat(outputDir)
	}
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	} else if !fi.IsDir() {
		HandleError(errors.Errorf("The provided path '%s' is not a directory", outputDir), cmd.opts.VerboseOutput)
	}

	// Clients
	batchclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}
	dataOps, err := aws.NewDataClient(
		aws.NewNerdalizeCredentials(batchclient, config.CurrentProject),
		nerd.GetCurrentUser().Region,
	)
	if err != nil {
		HandleError(errors.Wrap(err, "could not create aws dataops client"), cmd.opts.VerboseOutput)
	}
	dataclient := v1data.NewClient(dataOps)

	// Dataset
	ds, err := batchclient.DescribeDataset(config.CurrentProject, dataset)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}
	logrus.Infof("Downloading dataset with ID '%v'", ds.DatasetID)

	// Metadata
	metadata, err := dataclient.MetadataDownload(ds.Bucket, ds.DatasetRoot)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	// Index
	r, err := dataclient.Download(ds.Bucket, path.Join(ds.DatasetRoot, v1data.IndexObjectKey))
	if err != nil {
		HandleError(errors.Wrap(err, "failed to download chunk index"), cmd.opts.VerboseOutput)
	}

	// Progress bar
	progressCh := make(chan int64)
	progressBarDoneCh := make(chan struct{})
	if !cmd.opts.JSONOutput {
		go ProgressBar(metadata.Size, progressCh, progressBarDoneCh)
	} else {
		go func() {
			for _ = range progressCh {
			}
			progressBarDoneCh <- struct{}{}
		}()
	}

	// Untar
	doneCh := make(chan error)
	pr, pw := io.Pipe()
	go func() {
		uerr := untardir(outputDir, pr)
		pr.Close()
		doneCh <- uerr
	}()

	// Download
	err = dataclient.ChunkedDownload(v1data.NewIndexReader(r), pw, DownloadConcurrency, ds.Bucket, ds.DatasetRoot, progressCh)
	if err != nil {
		HandleError(errors.Wrapf(err, "failed to download project '%v'", dataset), cmd.opts.VerboseOutput)
	}
	close(progressCh)

	// Finish downloading
	err = pw.Close()
	if err != nil {
		HandleError(errors.Wrap(err, "failed to close chunked download pipe writer"), cmd.opts.VerboseOutput)
	}
	err = <-doneCh
	if err != nil {
		HandleError(errors.Wrapf(err, "failed to untar project '%v'", dataset), cmd.opts.VerboseOutput)
	}

	// Wait for progress bar to be flushed to screen
	<-progressBarDoneCh

	return nil
}

//untardir untars an archive from the reader to a directory on disk.
func untardir(dir string, r io.Reader) (err error) {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return errors.Wrap(err, "failed to read next tar header")
		}

		path := filepath.Join(dir, hdr.Name)
		err = os.MkdirAll(filepath.Dir(path), 0777)
		if err != nil {
			return errors.Wrap(err, "failed to create dirs")
		}

		f, err := safefile.Create(path, os.FileMode(hdr.Mode))
		if err != nil {
			return errors.Wrap(err, "failed to create tmp safe file")
		}

		defer f.Close()
		n, err := io.Copy(f, tr)
		if err != nil {
			return errors.Wrap(err, "failed to write file content to tmp file")
		}

		if n != hdr.Size {
			return errors.Errorf("unexpected nr of bytes written, wrote '%d' saw '%d' in tar hdr", n, hdr.Size)
		}

		err = f.Commit()
		if err != nil {
			return errors.Wrap(err, "failed to swap old file for tmp file")
		}

		err = os.Chtimes(path, time.Now(), hdr.ModTime)
		if err != nil {
			return errors.Wrap(err, "failed to change times of tmp file")
		}
	}

	return nil
}
