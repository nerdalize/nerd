package cmd

import (
	"bytes"
	"fmt"
	"reflect"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
	"github.com/posener/complete"
	"github.com/sirupsen/logrus"
)

var (
	//MessageNotEnoughArguments is shown when the user didn't provide enough arguments
	MessageNotEnoughArguments = "not enough arguments, see below for usage"

	//PlaceholderSynopsis is synopsis text when none is available
	PlaceholderSynopsis = "<synopsis>"

	//PlaceholderHelp is help when none is available
	PlaceholderHelp = "<help>"

	//PlaceholderUsage is shown when no specific implementation is available
	PlaceholderUsage = "<usage>"
)

type command struct {
	globalOpts struct {
		Debug bool `long:"debug" description:"show verbose debug information"`
	}

	flagParser *flags.Parser
	runFunc    func(args []string) error
	helpFunc   func() string
	usageFunc  func() string
	out        *Output
}

func createCommand(ui cli.Ui, runFunc func([]string) error, helpFunc func() string, usageFunc func() string, fgroup interface{}) *command {
	c := &command{
		usageFunc:  usageFunc,
		flagParser: flags.NewNamedParser(usageFunc(), flags.None),
		runFunc:    runFunc,
		helpFunc:   helpFunc,
		out:        NewOutput(ui),
	}

	_, err := c.flagParser.AddGroup("Options", "Options", fgroup)
	if err != nil {
		panic("failed to add option group: " + err.Error())
	}

	_, err = c.flagParser.AddGroup("Global Options", "Global Options", &c.globalOpts)
	if err != nil {
		panic("failed to add global option group: " + err.Error())
	}

	return c
}

func addFlagPredicts(fl complete.Flags, f *flags.Option) {
	if f.Field().Type.Kind() == reflect.Bool || f.Field().Type == reflect.SliceOf(reflect.TypeOf(true)) {
		fl["--"+f.LongName] = complete.PredictNothing
		if f.ShortName != 0 {
			fl[fmt.Sprintf("-%s", string(f.ShortName))] = complete.PredictNothing
		}

	} else {
		fl["--"+f.LongName] = complete.PredictAnything
		if f.ShortName != 0 {
			fl[fmt.Sprintf("-%s", string(f.ShortName))] = complete.PredictAnything
		}
	}
}

// AutocompleteFlags returns a mapping of supported flags
func (cmd *command) AutocompleteFlags() (fl complete.Flags) {
	fl = complete.Flags{}
	for _, g := range cmd.flagParser.Groups() {
		for _, f := range g.Options() {
			addFlagPredicts(fl, f)
		}
	}

	return fl
}

//Help shows extensive help
func (cmd *command) Help() string {
	buf := bytes.NewBuffer(nil)
	cmd.flagParser.WriteHelp(buf)
	return fmt.Sprintf(`
%s
%s`, cmd.helpFunc(), buf.String())
}

//Run runs the actual command
func (cmd *command) Run(args []string) int {
	remaining, err := cmd.flagParser.ParseArgs(args)
	if err != nil {
		return cmd.fail(err, "failed to parse flags(s)")
	}

	if err := cmd.runFunc(remaining); err != nil {
		switch cause := errors.Cause(err).(type) {
		case errShowUsage:
			return cmd.usage(cause)
		case errShowHelp:
			cmd.out.Output(cause.Error())
			return cli.RunResultHelp
		default:
			return cmd.fail(err, "error")
		}
	}
	return 0
}

//Logger returns the logger
func (cmd *command) Logger() *logrus.Logger {
	if cmd.globalOpts.Debug {
		return cmd.out.Logger(logrus.DebugLevel)
	}

	return cmd.out.Logger(logrus.ErrorLevel)
}

// AutocompleteArgs returns the argument predictor for this command.
func (cmd *command) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func (cmd *command) fail(err error, message string) int {
	cmd.out.Errorf("%v", errors.Wrap(err, message))
	return 255
}

func (cmd *command) usage(cause error) int {
	cmd.out.Output(fmt.Sprintf("%s\n\n%s", cause.Error(), cmd.usageFunc()))
	return 254
}
