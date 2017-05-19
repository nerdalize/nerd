package command

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

var errShowHelp = errors.New("show error")

func (cmd *command) setConfigLocation(loc string) {
	cmd.configLocation = loc
}

func setVerbose(verbose bool) {
	if verbose {
		logrus.SetFormatter(new(logrus.TextFormatter))
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func (cmd *command) setJSON(json bool) {
	cmd.jsonOutput = json
	if json {
		logrus.SetFormatter(new(logrus.JSONFormatter))
	}
}

func newCommand(title, synopsis, help string, opts interface{}) (*command, error) {
	cmd := &command{
		help:     help,
		synopsis: synopsis,
		parser:   flags.NewNamedParser(title, flags.None),
		ui: &cli.BasicUi{
			Reader: os.Stdin,
			Writer: os.Stderr,
		},
	}
	if opts != nil {
		_, err := cmd.parser.AddGroup("options", "options", opts)
		if err != nil {
			return nil, err
		}
	}
	confOpts := &ConfOpts{
		ConfigFile: cmd.setConfigLocation,
		OutputOpts: OutputOpts{
			VerboseOutput: setVerbose,
			JSONOutput:    cmd.setJSON,
		},
	}
	_, err := cmd.parser.AddGroup("output options", "output options", confOpts)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

//command is an abstract implementation for embedding in concrete commands and allows basic command functionality to be reused.
type command struct {
	help           string        //extended help message, show when --help a command
	synopsis       string        //short help message, shown on the command overview
	parser         *flags.Parser //option parser that will be used when parsing args
	ui             cli.Ui
	configLocation string
	jsonOutput     bool
	// conf     conf.ConfInterface
	// confOpts *ConfOpts
	// renderer Renderer
	runFunc func(args []string) error
}

//Will write help text for when a user uses --help, it automatically renders all option groups of the flags.Parser (augmented with default values). It will show an extended help message if it is not empty, else it shows the synopsis.
func (c *command) Help() string {
	buf := bytes.NewBuffer(nil)
	c.parser.WriteHelp(buf)

	txt := c.help
	if txt == "" {
		txt = c.Synopsis()
	}

	return fmt.Sprintf(`
%s

%s`, txt, buf.String())
}

//Short explanation of the command as passed in the struction initialization
func (c *command) Synopsis() string {
	return c.synopsis
}

//Run wraps a signature that allows returning an error type and parses the arguments for the flags package. If flag parsing fails it sets the exit code to 127, if the command implementation returns a non-nil error the exit code is 1
func (c *command) Run(args []string) int {
	if c.parser != nil {
		var err error
		args, err = c.parser.ParseArgs(args)
		if err != nil {
			return 127
		}
	}

	if err := c.runFunc(args); err != nil {
		if err == errShowHelp {
			return cli.RunResultHelp
		}
		c.ui.Error(err.Error())
		return 1
	}

	return 0
}
