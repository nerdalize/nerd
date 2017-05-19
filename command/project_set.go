package command

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
)

//ProjectSetOps describes command options
type ProjectSetOps struct {
	NerdOpts
}

//ProjectSet command
type ProjectSet struct {
	*command
	opts   *ProjectSetOps
	parser *flags.Parser
}

//ProjectSetFactory returns a factory method for the join command
func ProjectSetFactory() (cli.Command, error) {
	cmd := &ProjectSet{
		command: &command{
			help:     "",
			synopsis: "set current working project",
			parser:   flags.NewNamedParser("nerd project set <project-name>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &ProjectSetOps{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *ProjectSet) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	err = conf.WriteProject(args[0])
	if err != nil {
		HandleError(err)
	}

	return nil
}
