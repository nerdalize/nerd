package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

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
			synopsis: "push task data as input to cloud storage",
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

	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//DoRun is called by run and allows an error to be returned
func (cmd *Upload) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	SetLogSettings(cmd.opts.JSONOutput, cmd.opts.VerboseOutput)

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

	err = client.UploadDir(dataPath, ds.Root, &stdoutkw{}, 64)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	logrus.Infof("Uploaded data to dataset with ID: %v", ds.DatasetID)

	return nil
}
