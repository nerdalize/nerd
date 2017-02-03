package command

import (
	"bytes"
	"fmt"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//command is an abstract implementation for embedding in concrete commands and allows basic command functionality to be reused.
type command struct {
	help     string        //extended help message, show when --help a command
	synopsis string        //short help message, shown on the command overview
	parser   *flags.Parser //option parser that will be used when parsing args
	ui       cli.Ui
	runFunc  func(args []string) error
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
	a, err := c.parser.ParseArgs(args)
	if err != nil {
		return 127
	}

	if err := c.runFunc(a); err != nil {
		c.ui.Error(err.Error())
		return 1
	}

	return 0
}
