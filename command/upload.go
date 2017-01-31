package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd"
	"github.com/nerdalize/nerd/nerd/data"
)

//UploadOpts describes command options
type UploadOpts struct{}

//Upload command
type Upload struct {
	*command

	ui     cli.Ui
	opts   *UploadOpts
	parser *flags.Parser
}

//UploadFactory returns a factory method for the join command
func UploadFactory() func() (cmd cli.Command, err error) {
	cmd := &Upload{
		command: &command{
			help:     "",
			synopsis: "push task data as input to cloud storage",
			parser:   flags.NewNamedParser("nerd upload <dataset> <file-1> [<file-2> ... <file-n>]", flags.Default),
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
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	dataset := args[0]

	awsCreds, err := nerd.GetCurrentUser().GetAWSCredentials()
	if err != nil {
		return fmt.Errorf("could not get AWS credentials: %v", err)
	}

	client, err := data.NewClient(awsCreds)
	if err != nil {
		return fmt.Errorf("could not create data client: %v", err)
	}

	var files []string
	var errs []string

	for i := 1; i < len(args); i++ {
		f, err := os.Stat(args[i])
		if err != nil {
			errs = append(errs, fmt.Sprintf("argument '%v' is not a valid file or directory", args[i]))
			break
		}

		switch mode := f.Mode(); {
		case mode.IsDir():
			filepath.Walk(args[i], func(path string, f os.FileInfo, err error) error {
				if f.Mode().IsRegular() {
					files = append(files, path)
				}
				return nil
			})
		case mode.IsRegular():
			files = append(files, args[i])
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("the following error(s) occured when uploading to the dataset: %v", strings.Join(errs, ","))
	}

	err = client.UploadFiles(files, dataset, &stdoutkw{}, 64)
	if err != nil {
		return fmt.Errorf("could not upload files: %v", err)
	}

	// for i := 1; i < len(args); i++ {
	// 	f, err := os.Stat(args[i])
	// 	if err != nil {
	// 		errs = append(errs, fmt.Sprintf("argument '%v' is not a valid file or directory", args[i]))
	// 		break
	// 	}
	//
	// 	switch mode := f.Mode(); {
	// 	case mode.IsDir():
	// 		err = client.UploadDir(args[i], dataset)
	// 	case mode.IsRegular():
	// 		err = client.UploadFile(args[i], dataset)
	// 	}
	// 	if err != nil {
	// 		errs = append(errs, err.Error())
	// 	}
	// }

	return nil
}
