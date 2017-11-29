package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
	"github.com/posener/complete"
)

var (
	// errShowHelp can be returned by commands to show the commands help message next to the error
	errShowHelp = errors.New("show help")
)

var (
	//MessageNotEnoughArguments is shown when the user didn't provide enough arguments
	MessageNotEnoughArguments = "not enough arguments, see --help"

	//PlaceholderSynopsis is synopsis text when none is available
	PlaceholderSynopsis = "<synopsis>"

	//PlaceholderHelp is help when none is available
	PlaceholderHelp = "<help>"

	//PlaceholderUsage is sown when no specific implementation is available
	PlaceholderUsage = "<usage>"
)

type command struct {
	flagParser *flags.Parser
	runFunc    func(args []string) error
	helpFunc   func() string
	logs       *log.Logger
}

func createCommand(runFunc func([]string) error, helpFunc func() string, usageFunc func() string, fgroup interface{}) *command {
	c := &command{
		flags.NewNamedParser(usageFunc(), flags.None),
		runFunc,
		helpFunc,
		log.New(os.Stderr, "", 0),
	}

	_, err := c.flagParser.AddGroup("Options", "Options", fgroup)
	if err != nil {
		panic("failed to add option group: " + err.Error())
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
		return fail(err, "failed to parse flags(s)")
	}

	if err := cmd.runFunc(remaining); err != nil {
		if err == errShowHelp {
			return cli.RunResultHelp
		}

		return fail(err, "failed to run")
	}

	return 0
}

//Close can be used on defer and shows any errors through the logs
func (cmd *command) Close(name string, cl io.Closer) {
	err := cl.Close()
	if err != nil {
		cmd.logs.Printf("[ERRO] failure while shutting down '%s': %v", name, err)
	}
}

// AutocompleteArgs returns the argument predictor for this command.
func (cmd *command) AutocompleteArgs() complete.Predictor {
	return complete.PredictNothing
}

func fail(err error, message string) int {
	fmt.Fprintf(os.Stderr, "[ERRO] %v\n", errors.Wrap(err, message))
	return 255
}
