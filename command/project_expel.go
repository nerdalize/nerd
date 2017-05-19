package command

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
)

//ProjectExpelOps describes command options
type ProjectExpelOps struct {
	NerdOpts
}

//ProjectExpel command
type ProjectExpel struct {
	*command
	opts   *ProjectExpelOps
	parser *flags.Parser
}

//ProjectExpelFactory returns a factory method for the join command
func ProjectExpelFactory() (cli.Command, error) {
	cmd := &ProjectExpel{
		command: &command{
			help:     "",
			synopsis: "expel the current project from its compute cluster",
			parser:   flags.NewNamedParser("nerd project expel", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &ProjectExpelOps{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *ProjectExpel) DoRun(args []string) (err error) {
	config, err := conf.Read()
	if err != nil {
		HandleError(err)
	}

	bclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err)
	}

	out, err := bclient.ExpelProject(config.CurrentProject.Name)
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Placement removed: %v", out)
	return nil
}
