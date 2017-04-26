package command

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd"
	"github.com/nerdalize/nerd/nerd/aws"
	v1batch "github.com/nerdalize/nerd/nerd/client/batch/v1"
	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
	v1data "github.com/nerdalize/nerd/nerd/client/data/v1"
	v1datapayload "github.com/nerdalize/nerd/nerd/client/data/v1/payload"
	"github.com/nerdalize/nerd/nerd/conf"
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
}

//Upload command
type Upload struct {
	*command

	opts   *UploadOpts
	parser *flags.Parser
}

//UploadFactory returns a factory method for the join command
func UploadFactory() func() (cmd cli.Command, err error) {
	cmd := &Upload{
		command: &command{
			help:     "",
			synopsis: "Upload a dataset to cloud storage.\nOptionally you can specify a dataset-ID to append files to that dataset. This is also useful to continue an upload in case a previous try failed.",
			parser:   flags.NewNamedParser("nerd upload <path> [dataset-ID]", flags.Default),
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

	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//DoRun is called by run and allows an error to be returned
func (cmd *Upload) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	dataPath := args[0]
	datasetID := ""
	if len(args) == 2 {
		datasetID = args[1]
	}

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
		aws.NewNerdalizeCredentials(batchclient, config.CurrentProject),
		nerd.GetCurrentUser().Region,
	)
	if err != nil {
		HandleError(errors.Wrap(err, "could not create aws dataops client"), cmd.opts.VerboseOutput)
	}
	dataclient := v1data.NewClient(dataOps)

	// Dataset
	ds, err := getDataset(batchclient, config.CurrentProject, datasetID)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}
	logrus.Infof("Uploading dataset with ID '%v'", ds.DatasetID)
	err = ioutil.WriteFile(path.Join(dataPath, DatasetFilename), []byte(ds.DatasetID), DatasetPermissions)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	// Index
	indexr, indexw := io.Pipe()
	indexDoneCh := make(chan error)
	go func() {
		b, err := ioutil.ReadAll(indexr)
		if err != nil {
			indexDoneCh <- errors.Wrap(err, "failed to read keys")
			return
		}
		indexDoneCh <- dataclient.Upload(ds.Bucket, path.Join(ds.Root, v1data.IndexObjectKey), bytes.NewReader(b))
	}()
	iw := v1data.NewIndexWriter(indexw)

	// Progress
	size, err := totalTarSize(dataPath)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}
	progressCh := make(chan int64)
	progressBarDoneCh := make(chan struct{})
	if !cmd.opts.JSONOutput {
		go ProgressBar(size, progressCh, progressBarDoneCh)
	} else {
		go func() {
			for _ = range progressCh {
			}
			progressBarDoneCh <- struct{}{}
		}()
	}

	// Uploading
	doneCh := make(chan error)
	pr, pw := io.Pipe()
	go func() {
		defer close(progressCh)
		err := dataclient.ChunkedUpload(NewChunker(v1data.UploadPolynomal, pr), iw, UploadConcurrency, ds.Bucket, ds.Root, progressCh)
		pr.Close()
		doneCh <- err
	}()

	// Tarring
	err = tardir(dataPath, pw)
	if err != nil && errors.Cause(err) != io.ErrClosedPipe {
		HandleError(errors.Wrapf(err, "failed to tar '%s'", dataPath), cmd.opts.VerboseOutput)
	}

	// Finish uploading
	err = pw.Close()
	if err != nil {
		HandleError(errors.Wrap(err, "failed to close chunked upload pipe writer"), cmd.opts.VerboseOutput)
	}
	err = <-doneCh
	if err != nil {
		HandleError(errors.Wrapf(err, "failed to upload '%s'", dataPath), cmd.opts.VerboseOutput)
	}

	// Finish uploading index
	err = iw.Close()
	if err != nil {
		HandleError(errors.Wrap(err, "failed to close index pipe writer"), cmd.opts.VerboseOutput)
	}
	err = <-indexDoneCh
	if err != nil {
		HandleError(errors.Wrap(err, "failed to upload index file"), cmd.opts.VerboseOutput)
	}

	// Wait for progress bar to be flushed to screen
	<-progressBarDoneCh

	// Metadata
	metadata, err := getMetadata(dataclient, ds.Bucket, ds.Root, size)
	if err != nil {
		HandleError(errors.Wrapf(err, "failed to get metadata for dataset '%v'", datasetID), cmd.opts.VerboseOutput)
	}
	err = dataclient.MetadataUpload(ds.Bucket, ds.Root, metadata)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

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
		defer f.Close()
		if err != nil {
			return errors.Wrapf(err, "failed to open file '%s'", rel)
		}

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

//getDataset returns a payload.Dataset object. If datasetID is set it will try to read an existing dataset, if datasetID is empty a new dataset will be created.
func getDataset(client *v1batch.Client, projectID, datasetID string) (*v1payload.Dataset, error) {
	if datasetID == "" {
		dsc, err := client.CreateDataset(projectID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create dataset")
		}
		return &dsc.Dataset, nil
	}
	dsg, err := client.GetDataset(projectID, datasetID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve dataset")
	}
	return &dsg.Dataset, nil
}

func getMetadata(client *v1data.Client, bucket, root string, size int64) (*v1datapayload.Metadata, error) {
	exists, err := client.MetadataExists(bucket, root)
	if err != nil {
		return nil, errors.Wrap(err, "failed to check if metadata exists")
	}
	if !exists {
		return &v1datapayload.Metadata{
			Size:    size,
			Created: time.Now(),
			Updated: time.Now(),
		}, nil
	}
	metadata, err := client.MetadataDownload(bucket, root)
	metadata.Size = size
	metadata.Updated = time.Now()
	return metadata, nil
}
