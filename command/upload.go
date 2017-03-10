package command

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/aws"
	"github.com/nerdalize/nerd/nerd/conf"
)

//UploadOpts describes command options
type UploadOpts struct {
	*NerdOpts
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

	conf.SetLocation(cmd.opts.ConfigFile)

	path := args[0]

	nerdclient, err := NewClient(cmd.ui)
	if err != nil {
		return HandleError(HandleClientError(err, cmd.opts.VerboseOutput), cmd.opts.VerboseOutput)
	}
	ds, err := nerdclient.CreateDataset()
	if err != nil {
		return HandleError(HandleClientError(err, cmd.opts.VerboseOutput), cmd.opts.VerboseOutput)
	}

	client, err := aws.NewDataClient(&aws.DataClientConfig{
		Credentials: aws.NewNerdalizeCredentials(nerdclient),
		Bucket:      ds.Bucket,
	})
	if err != nil {
		return fmt.Errorf("could not create data client: %v", err)
	}

	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("argument '%v' is not a valid file or directory", path)
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		err = client.UploadDir(path, ds.Root, &stdoutkw{}, 64)
	case mode.IsRegular():
		err = client.UploadFile(path, path, ds.Root)
	}

	if err != nil {
		return err
	}

	fmt.Printf("\nUploaded data to dataset with ID: %v\n", ds.DatasetID)

	return nil
}
