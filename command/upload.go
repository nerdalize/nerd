package command

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	pb "gopkg.in/cheggaaa/pb.v1"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/aws"
	"github.com/pkg/errors"
)

const (
	//DatasetFilename is the filename of the file that contains the dataset ID in the data folder.
	DatasetFilename = ".dataset"
	//DatasetPermissions are the permissions for DatasetFilename
	DatasetPermissions = 0644
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

	nerdclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}
	ds, err := nerdclient.CreateDataset()
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

	fi, err := os.Stat(dataPath)
	if err != nil {
		HandleError(errors.Errorf("argument '%v' is not a valid file or directory", dataPath), cmd.opts.VerboseOutput)
	} else if !fi.IsDir() {
		HandleError(errors.Errorf("provided path '%s' is not a directory", dataPath), cmd.opts.VerboseOutput)
	}

	err = ioutil.WriteFile(path.Join(dataPath, DatasetFilename), []byte(ds.DatasetID), DatasetPermissions)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	keyrw := KeyReadWriter()
	if cmd.opts.JSONOutput {
		doneCh := make(chan error)
		progressCh := make(chan int64)
		pr, pw := io.Pipe()
		go func() {
			doneCh <- client.ChunkedUpload(pr, keyrw, 64, ds.Root, progressCh)
		}()

		err = Tar(dataPath, pw)
		if err != nil {
			return fmt.Errorf("failed to tar '%s': %v", dataPath, err)
		}

		pw.Close()
		err = <-doneCh
		if err != nil {
			return fmt.Errorf("failed to upload: %v", err)
		}
	} else {
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

		err = Tar(dataPath, pw)
		if err != nil {
			return fmt.Errorf("failed to tar '%s': %v", dataPath, err)
		}

		pw.Close()
		cr := <-doneCh
		if cr.err != nil {
			return fmt.Errorf("failed to upload: %v", err)
		}
		bar := pb.New64(cr.total).Start()
		doneCh2 := make(chan error)
		progressCh := make(chan int64)
		pr, pw = io.Pipe()
		go func() {
			doneCh2 <- client.ChunkedUpload(pr, keyrw, 64, ds.Root, progressCh)
		}()

		go func() {
			for elem := range progressCh {
				bar.Add64(elem)
			}
		}()

		err = Tar(dataPath, pw)
		if err != nil {
			return fmt.Errorf("failed to tar '%s': %v", dataPath, err)
		}

		pw.Close()
		err = <-doneCh2
		if err != nil {
			return fmt.Errorf("failed to upload: %v", err)
		}
	}
	//push index file
	var ks []string
	for k, e := keyrw.Read(); e == nil; k, e = keyrw.Read() {
		ks = append(ks, k)
	}
	err = client.Upload(path.Join(ds.Root, "index"), strings.NewReader(strings.Join(ks, "\n")))
	if err != nil {
		return errors.Wrap(err, "failed to upload index file")
	}

	logrus.Infof("Uploaded data to dataset with ID: %v", ds.DatasetID)

	return nil
}

//Tar archives the given directory and writes bytes to
func Tar(dir string, w io.Writer) (err error) {
	tw := tar.NewWriter(w)
	err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if fi.Mode().IsDir() {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("failed to determine path '%s' relative to '%s': %v", path, dir, err)
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file '%s': %v", rel, err)
		}

		err = tw.WriteHeader(&tar.Header{
			Name:    rel,
			Mode:    int64(fi.Mode()),
			ModTime: fi.ModTime(),
			Size:    fi.Size(),
		})
		if err != nil {
			return fmt.Errorf("failed to write tar header for '%s': %v", rel, err)
		}

		defer f.Close()
		n, err := io.Copy(tw, f)
		if err != nil {
			return fmt.Errorf("failed to write tar file for '%s': %v", rel, err)
		}

		if n != fi.Size() {
			return fmt.Errorf("unexpected nr of bytes written to tar, saw '%d' on-disk but only wrote '%d', is directory '%s' in use?", fi.Size(), n, dir)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk dir '%s': %v", dir, err)
	}

	if err = tw.Close(); err != nil {
		return fmt.Errorf("failed to write remaining data: %v", err)
	}

	return nil
}

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
