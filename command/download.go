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
	"github.com/nerdalize/nerd/nerd/aws"
	"github.com/nerdalize/nerd/nerd/data"
	"github.com/pkg/errors"
)

const (
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

//DownloadFactory returns a factory method for the join command
func DownloadFactory() func() (cmd cli.Command, err error) {
	cmd := &Download{
		command: &command{
			help:     "",
			synopsis: "Download a dataset from cloud storage.",
			parser:   flags.NewNamedParser("nerd download <dataset> <output-dir>", flags.Default),
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

	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//DoRun is called by run and allows an error to be returned
func (cmd *Download) DoRun(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments, see --help")
	}
	dataset := args[0]
	outputDir := args[1]

	fi, err := os.Stat(outputDir)
	// create directory if it does not exist yet.
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

	nerdclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}
	ds, err := nerdclient.GetDataset(dataset)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	client, err := aws.NewDataClient(&aws.DataClientConfig{
		Credentials: aws.NewNerdalizeCredentials(nerdclient),
		Bucket:      ds.Bucket,
	})
	if err != nil {
		HandleError(errors.Wrap(err, "could not create data client"), cmd.opts.VerboseOutput)
	}

	r, err := client.Download(path.Join(ds.Root, data.MetadataObjectKey))
	if err != nil {
		HandleError(errors.Wrap(err, "failed to download metadata"), cmd.opts.VerboseOutput)
	}
	defer r.Close()
	metadata, err := data.NewMetadataFromReader(r)
	if err != nil {
		HandleError(errors.Wrap(err, "failed to read metadata"), cmd.opts.VerboseOutput)
	}

	logrus.Infof("Downloading dataset with ID '%v'", ds.DatasetID)

	doneCh := make(chan error)
	progressCh := make(chan int64)
	progressBarDoneCh := make(chan struct{})
	pr, pw := io.Pipe()
	go func() {
		err := untardir(outputDir, pr)
		pr.Close()
		doneCh <- err
	}()
	if !cmd.opts.JSONOutput {
		go ProgressBar(metadata.Header.Size, progressCh, progressBarDoneCh)
	} else {
		go func() {
			for _ = range progressCh {
			}
			progressBarDoneCh <- struct{}{}
		}()
	}
	err = client.ChunkedDownload(metadata, pw, DownloadConcurrency, ds.Root, progressCh)
	close(progressCh)
	if err != nil {
		HandleError(errors.Wrapf(err, "failed to download project '%v'", dataset), cmd.opts.VerboseOutput)
	}

	pw.Close()
	err = <-doneCh
	<-progressBarDoneCh
	if err != nil {
		HandleError(errors.Wrapf(err, "failed to untar project '%v'", dataset), cmd.opts.VerboseOutput)
	}

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
