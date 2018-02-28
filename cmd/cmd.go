package cmd

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/cheggaaa/pb"
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

func createCommand(ui cli.Ui, runFunc func([]string) error, helpFunc func() string, usageFunc func() string, fgroup interface{}, opts flags.Options) *command {
	c := &command{
		flagParser: flags.NewNamedParser(usageFunc(), opts),
		runFunc:    runFunc,
		helpFunc:   helpFunc,
		usageFunc:  usageFunc,
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

// Options returns the available options of a command
func (cmd *command) Options() *flags.Parser {
	return cmd.flagParser
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
			return cmd.fail(err, "Error")
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

//implements the transfer reporter such that it shows archiving progress
type progressBarReporter struct {
	uarch *pb.ProgressBar
	arch  *pb.ProgressBar
	upl   *pb.ProgressBar
	dwn   *pb.ProgressBar
}

func (r *progressBarReporter) HandledKey(key string) {}

func (r *progressBarReporter) StartArchivingProgress(label string, total int64) func(int64) {
	if total == 0 {
		return func(n int64) {}
	}
	r.arch = pb.New(int(total)).SetUnits(pb.U_BYTES_DEC)
	r.arch.Prefix(fmt.Sprintf("Archiving (Step 1/2):")) //@TODO with debug flag show temp file
	r.arch.Start()

	return func(n int64) {
		r.arch.Add64(n)
	}
}

func (r *progressBarReporter) StartUploadProgress(label string, total int64, rr io.Reader) io.Reader {
	r.upl = pb.New(int(total)).SetUnits(pb.U_BYTES_DEC)
	if r.arch != nil {
		r.upl.Prefix("Uploading (Step 2/2):") //@TODO with debug flag show key for uploading
	} else {
		r.upl.Prefix("Uploading:")
	}
	r.upl.Start()

	return r.upl.NewProxyReader(rr)
}

func (r *progressBarReporter) StopUploadProgress() {
	r.upl.Finish()
}

func (r *progressBarReporter) StopArchivingProgress() {
	if r.arch == nil {
		return
	}
	r.arch.Finish()
}

func (r *progressBarReporter) StartDownloadProgress(label string, total int64) io.Writer {
	r.dwn = pb.New(int(total)).SetUnits(pb.U_BYTES_DEC)
	r.dwn.Prefix(fmt.Sprintf("Downloading (Step 1/2):")) //@TODO with debug flag show key
	r.dwn.Start()

	return r.dwn
}

func (r *progressBarReporter) StopDownloadProgress() {
	r.dwn.Finish()
}

func (r *progressBarReporter) StartUnarchivingProgress(label string, total int64, rr io.Reader) io.Reader {
	r.uarch = pb.New(int(total)).SetUnits(pb.U_BYTES_DEC)
	r.uarch.Prefix(fmt.Sprintf("Unarchiving (Step 2/2):")) //@TODO with debug flag show temp file
	r.uarch.Start()

	return r.uarch.NewProxyReader(rr)
}

func (r *progressBarReporter) StopUnarchivingProgress() {
	r.uarch.Finish()
}
