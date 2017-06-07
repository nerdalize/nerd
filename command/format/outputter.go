package format

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/hashicorp/logutils"
	"github.com/pkg/errors"
)

//OutputType is one of prett, raw, or json
type OutputType string

//DecMap maps an OutputType to a Decorator
type DecMap map[OutputType]Decorator

//Decorator decorates a value and writes to out
type Decorator interface {
	Decorate(out io.Writer) error
}

const (
	//OutputTypePretty is used for pretty printing
	OutputTypePretty = "pretty"
	//OutputTypeRaw is used for raw output (nice for unix piping)
	OutputTypeRaw = "raw"
	//OutputTypeJSON is used for JSON output
	OutputTypeJSON = "json"
)

var (
	debugFlags = log.LstdFlags | log.Lshortfile
)

//Outputter is responsible for all output
type Outputter struct {
	debug      bool
	outputType OutputType
	outw       io.Writer
	errw       io.Writer
	logfile    io.WriteCloser
	Logger     *log.Logger
}

//NewOutputter creates a new Outputter that writes to Stdout and Stderr
func NewOutputter(outw, errw io.Writer, logger *log.Logger) *Outputter {
	o := &Outputter{
		outw:   outw,
		errw:   errw,
		Logger: logger,
	}
	o.setFilter()
	return o
}

//setFilter sets the logger output filter
func (o *Outputter) setFilter() {
	if o.Logger == nil {
		panic("no logger specified")
	}
	logLevel := "INFO"
	if o.debug {
		logLevel = "DEBUG"
	}
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "WARN", "ERROR", "INFO"},
		MinLevel: logutils.LogLevel(logLevel),
		Writer:   o.errw,
	}
	if o.logfile != nil {
		o.Logger.SetOutput(io.MultiWriter(o.logfile, filter))
	} else {
		o.Logger.SetOutput(filter)
	}
}

//Close closes the log file
func (o *Outputter) Close() error {
	if o.logfile != nil {
		return o.logfile.Close()
	}
	return nil
}

//ErrW returns the err writer
func (o *Outputter) ErrW() io.Writer {
	return o.errw
}

//SetOutputType sets the output type
func (o *Outputter) SetOutputType(ot OutputType) {
	o.outputType = ot
}

//SetDebug sets debug outputting
func (o *Outputter) SetDebug(debug bool) {
	o.debug = debug
	o.setFilter()
	if debug == true {
		o.Logger.SetFlags(debugFlags)
	}
}

//SetLogToDisk sets a logfile to write to
func (o *Outputter) SetLogToDisk(location string) error {
	f, err := os.OpenFile(location, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to open log file")
	}
	o.logfile = f
	o.setFilter()
	return nil
}

//multi returns a MultiWriter if the logfile is set
func (o *Outputter) multi(w io.Writer) io.Writer {
	if o.logfile == nil {
		return w
	}
	return io.MultiWriter(w, o.logfile)
}

//Output outputs using the right decorator
func (o *Outputter) Output(d DecMap) {
	deco, ok := d[o.outputType]
	if !ok {
		deco = NotImplDecorator(o.outputType)
	}
	err := deco.Decorate(o.multi(o.outw))
	if err != nil {
		o.WriteError(errors.Wrap(err, "failed to decorate output"))
	}
}

//WriteError writes an error to errw
func (o *Outputter) WriteError(err error) {
	if o.Logger == nil {
		panic("no logger specified")
	}
	if errors.Cause(err) != nil { // when there's are more than 1 message on the message stack, only print the top one for user friendlyness.
		o.Logger.Println(strings.Replace(err.Error(), ": "+errorCauser(errorCauser(err)).Error(), "", 1))
	} else {
		o.Logger.Println(err)
	}
	o.Logger.Printf("[DEBUG] Underlying error: %+v\n", err)
}

//errorCauser returns the error that is one level up in the error chain.
func errorCauser(err error) error {
	type causer interface {
		Cause() error
	}

	if err2, ok := err.(causer); ok {
		err = err2.Cause()
	}
	return err
}
