package command

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
)

//ProjectPlaceOps describes command options
type ProjectPlaceOps struct {
	NerdOpts
}

//ProjectPlace command
type ProjectPlace struct {
	*command
	opts   *ProjectPlaceOps
	parser *flags.Parser
}

//ProjectPlaceFactory returns a factory method for the join command
func ProjectPlaceFactory() (cli.Command, error) {
	cmd := &ProjectPlace{
		command: &command{
			help:     "",
			synopsis: "place the current project on a compute cluster",
			parser:   flags.NewNamedParser("nerd project place <host> <token>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &ProjectPlaceOps{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *ProjectPlace) DoRun(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	config, err := conf.Read()
	if err != nil {
		HandleError(err)
	}

	bclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err)
	}

	out, err := bclient.PlaceProject(config.CurrentProject.Name, args[0], args[1], "") //@TODO allow self signed certificate
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Placement created: %v", out)
	return nil
}
