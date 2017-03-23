package command

import (
	"archive/tar"
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dchest/safefile"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/aws"
	"github.com/pkg/errors"
)

const (
	OutputDirPermissions = 0755
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

	// TODO: index magic word
	r, err := client.Download(path.Join(ds.Root, "index"))
	defer r.Close()
	keyrw := NewlineKeyReader(r, ds.Root)

	doneCh := make(chan error)
	pr, pw := io.Pipe()
	go func() {
		doneCh <- untardir(outputDir, pr)
	}()
	err = client.ChunkedDownload(keyrw, pw, 64)
	if err != nil {
		return fmt.Errorf("failed to download project: %v", err)
	}

	return nil
}

//OverwriteHandlerUserPrompt is a handler that checks wether a file should be overwritten by asking the user over Stdin.
func OverwriteHandlerUserPrompt(ui cli.Ui) func(string) bool {
	return func(file string) bool {
		question := fmt.Sprintf("The file '%v' already exists. Do you want to overwrite it? [Y/n]", file)
		ans, err := ui.Ask(question)
		if err != nil {
			ui.Info(fmt.Sprintf("Failed to read your answer, '%v' will be skipped", file))
			return false
		}
		if ans == "n" {
			return false
		}
		return true
	}
}

//AlwaysOverwriteHandler is a handler that tells to always overwrite a file.
func AlwaysOverwriteHandler(file string) bool {
	return true
}

func untardir(dir string, r io.Reader) (err error) {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return fmt.Errorf("failed to read next tar header: %v", err)
		}

		path := filepath.Join(dir, hdr.Name)
		err = os.MkdirAll(filepath.Dir(path), 0777)
		if err != nil {
			return fmt.Errorf("failed to create dirs: %v", err)
		}

		f, err := safefile.Create(path, os.FileMode(hdr.Mode))
		if err != nil {
			return fmt.Errorf("failed to create tmp safe file: %v", err)
		}

		defer f.Close()
		n, err := io.Copy(f, tr)
		if err != nil {
			return fmt.Errorf("failed to write file content to tmp file: %v", err)
		}

		if n != hdr.Size {
			return fmt.Errorf("unexpected nr of bytes written, wrote '%d' saw '%d' in tar hdr", n, hdr.Size)
		}

		err = f.Commit()
		if err != nil {
			return fmt.Errorf("failed to swap old file for tmp file: %v", err)
		}

		err = os.Chtimes(path, time.Now(), hdr.ModTime)
		if err != nil {
			return fmt.Errorf("failed to change times of tmp file: %v", err)
		}
	}

	return nil
}

type newlineKeyReader struct {
	*sync.Mutex
	r    *bufio.Reader
	root string
}

// TODO: Abstract root
func NewlineKeyReader(r io.Reader, root string) *newlineKeyReader {
	return &newlineKeyReader{
		Mutex: new(sync.Mutex),
		r:     bufio.NewReader(r),
		root:  root,
	}
}

func (kw *newlineKeyReader) Read() (k string, err error) {
	kw.Lock()
	defer kw.Unlock()
	line, err := kw.r.ReadString('\n')
	if err != nil {
		return "", errors.Wrap(err, "failed to read key from input stream")
	}
	return path.Join(kw.root, strings.Replace(line, "\n", "", 1)), nil
}
