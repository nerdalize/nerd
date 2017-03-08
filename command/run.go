package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
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
			synopsis: "create a new compute task for a dataset",
			parser:   flags.NewNamedParser("nerd run <image> <dataset>", flags.Default),
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
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	conf.SetLocation(cmd.opts.ConfigFile)

	env := make(map[string]string)
	for i, e := range cmd.opts.Environment {
		split := strings.Split(e, "=")
		if len(split) != 2 {
			return errors.Errorf("Environment variable %v (%v) is in the wrong format. Please specify environment flag as '-e [KEY]=[VALUE]'", (i + 1), e)
		}
		env[split[0]] = split[1]
	}

	client, err := NewClient(cmd.ui)
	if err != nil {
		return HandleError(HandleClientError(err, cmd.opts.VerboseOutput), cmd.opts.VerboseOutput)
	}

	task, err := client.CreateTask(args[0], args[1], env)
	if err != nil {
		return HandleError(HandleClientError(err, cmd.opts.VerboseOutput), cmd.opts.VerboseOutput)
	}
	fmt.Printf("Created task with ID %v\n", task.TaskID)
	return nil
}
