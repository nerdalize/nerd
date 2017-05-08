package command

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/aws"
	"github.com/nerdalize/nerd/nerd/conf"
	v1datatransfer "github.com/nerdalize/nerd/nerd/service/datatransfer/v1"
	"github.com/pkg/errors"
)

const (
	//DatasetFilename is the filename of the file that contains the dataset ID in the data folder.
	DatasetFilename = ".dataset"
	//DatasetPermissions are the permissions for DatasetFilename
	DatasetPermissions = 0644
	//UploadConcurrency is the amount of concurrent upload threads.
	UploadConcurrency = 64
)

//UploadOpts describes command options
type UploadOpts struct {
	NerdOpts
	Tag string `long:"tag" default:"" default-mask:"" description:"use a tag to logically group datasets"`
}

//Upload command
type Upload struct {
	*command

	opts   *UploadOpts
	parser *flags.Parser
}

//DatasetUploadFactory returns a factory method for the join command
func DatasetUploadFactory() (cli.Command, error) {
	cmd := &Upload{
		command: &command{
			help:     "",
			synopsis: "upload data to the cloud and create a new dataset",
			parser:   flags.NewNamedParser("nerd upload <path>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &UploadOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Upload) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	dataPath := args[0]

	fi, err := os.Stat(dataPath)
	if err != nil {
		HandleError(errors.Errorf("argument '%v' is not a valid file or directory", dataPath), cmd.opts.VerboseOutput)
	} else if !fi.IsDir() {
		HandleError(errors.Errorf("provided path '%s' is not a directory", dataPath), cmd.opts.VerboseOutput)
	}

	// Config
	config, err := conf.Read()
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	// Clients
	batchclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}
	dataOps, err := aws.NewDataClient(
		aws.NewNerdalizeCredentials(batchclient, config.CurrentProject.Name),
		config.CurrentProject.AWSRegion,
	)
	if err != nil {
		HandleError(errors.Wrap(err, "could not create aws dataops client"), cmd.opts.VerboseOutput)
	}

	progressCh := make(chan int64)
	progressBarDoneCh := make(chan struct{})
	size, err := totalTarSize(dataPath)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}
	go ProgressBar(size, progressCh, progressBarDoneCh)

	err = v1datatransfer.Upload(v1datatransfer.UploadConfig{
		BatchClient: batchclient,
		DataOps:     dataOps,
		LocalDir:    dataPath,
		ProjectID:   config.CurrentProject.Name,
		Tag:         cmd.opts.Tag,
		Concurrency: 64,
		ProgressCh:  progressCh,
	})
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}
	<-progressBarDoneCh
	return nil
}

//tardir archives the given directory and writes bytes to w.
func tardir(dir string, w io.Writer) (err error) {
	tw := tar.NewWriter(w)
	err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if fi.Mode().IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return errors.Wrapf(err, "failed to determine path '%s' relative to '%s'", path, dir)
		}

		f, err := os.Open(path)
		if err != nil {
			return errors.Wrapf(err, "failed to open file '%s'", rel)
		}
		defer f.Close()

		err = tw.WriteHeader(&tar.Header{
			Name:    rel,
			Mode:    int64(fi.Mode()),
			ModTime: fi.ModTime(),
			Size:    fi.Size(),
		})
		if err != nil {
			return errors.Wrapf(err, "failed to write tar header for '%s'", rel)
		}

		n, err := io.Copy(tw, f)
		if err != nil {
			return errors.Wrapf(err, "failed to write tar file for '%s'", rel)
		}

		if n != fi.Size() {
			return errors.Errorf("unexpected nr of bytes written to tar, saw '%d' on-disk but only wrote '%d', is directory '%s' in use?", fi.Size(), n, dir)
		}

		return nil
	})

	if err != nil {
		return errors.Wrapf(err, "failed to walk dir '%s'", dir)
	}

	if err = tw.Close(); err != nil {
		return errors.Wrap(err, "failed to write remaining data")
	}

	return nil
}

//countBytes counts all bytes from a reader.
func countBytes(r io.Reader) (int64, error) {
	var total int64
	buf := make([]byte, 512*1024)
	for {
		n, err := io.ReadFull(r, buf)
		if err == io.ErrUnexpectedEOF {
			err = nil
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, errors.Wrap(err, "failed to read part of tar")
		}
		total = total + int64(n)
	}
	return total, nil
}

//totalTarSize calculates the total size in bytes of the archived version of a directory on disk.
func totalTarSize(dataPath string) (int64, error) {
	type countResult struct {
		total int64
		err   error
	}
	doneCh := make(chan countResult)
	pr, pw := io.Pipe()
	go func() {
		total, err := countBytes(pr)
		doneCh <- countResult{total, err}
	}()

	err := tardir(dataPath, pw)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to tar '%s'", dataPath)
	}

	pw.Close()
	cr := <-doneCh
	if cr.err != nil {
		return 0, errors.Wrapf(err, "failed to count total disk size of '%v'", dataPath)
	}
	return cr.total, nil
}
