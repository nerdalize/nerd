package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//RunOpts describes command options
type RunOpts struct {
	*NerdOpts

	Environment []string `short:"e" description:"container environment variables"`
}

//Run command
type Run struct {
	*command

	opts   *RunOpts
	parser *flags.Parser
}

//RunFactory returns a factory method for the join command
func RunFactory() func() (cmd cli.Command, err error) {
	cmd := &Run{
		command: &command{
			help:     "",
			synopsis: "Run a new compute task.\nOptionally you can give a dataset-ID as the second argument. This dataset-ID will be available as the NERD_DATASET_INPUT env variable inside the container.",
			parser:   flags.NewNamedParser("nerd run <image> [dataset-ID]", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &RunOpts{},
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
func (cmd *Run) DoRun(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	dataset := ""
	if len(args) == 2 {
		dataset = args[1]
	}

	env := make(map[string]string)
	for i, e := range cmd.opts.Environment {
		split := strings.Split(e, "=")
		if len(split) != 2 {
			HandleError(errors.Errorf("Environment variable %v (%v) is in the wrong format. Please specify environment flag as '-e [KEY]=[VALUE]'", (i+1), e), cmd.opts.VerboseOutput)
		}
		env[split[0]] = split[1]
	}

	client, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	task, err := client.CreateTask(args[0], dataset, env)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}
	logrus.Infof("Created task with ID %v", task.TaskID)
	return nil
}
