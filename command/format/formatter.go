package format

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type OutputType string
type DecMap map[OutputType]Decorator

type Decorator interface {
	Decorate(out io.Writer) error
}

const (
	OutputTypePretty = "pretty"
	OutputTypeRaw    = "raw"
	OutputTypeJSON   = "json"
)

type Outputter struct {
	verbose    bool
	outputType OutputType
	outw       io.Writer
	errw       io.Writer
	logfile    io.WriteCloser
}

func NewOutputter() *Outputter {
	return &Outputter{
		outw: os.Stderr,
		errw: os.Stdout,
	}
}

func (o *Outputter) Close() error {
	if o.logfile != nil {
		return o.logfile.Close()
	}
	return nil
}

func (o *Outputter) SetOutputType(ot OutputType) {
	o.outputType = ot
}

func (o *Outputter) SetVerbose(v bool) {
	o.verbose = v
}

func (o *Outputter) SetLogToDisk(location string) error {
	f, err := os.OpenFile(location, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return errors.Wrap(err, "failed to open log file")
	}
	o.logfile = f
	return nil
}

func (o *Outputter) multi(w io.Writer) io.Writer {
	if o.logfile == nil {
		return w
	}
	return io.MultiWriter(w, o.logfile)
}

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

func (o *Outputter) WriteError(err error) {
	if errors.Cause(err) != nil { // when there's are more than 1 message on the message stack, only print the top one for user friendlyness.
		o.Info(strings.Replace(err.Error(), ": "+errorCauser(errorCauser(err)).Error(), "", 1))
	} else {
		o.Info(err)
	}
	o.Debugf("Underlying error: %+v", err)
}

func (o *Outputter) Info(a ...interface{}) {
	fmt.Fprint(o.multi(o.outw), a)
}

func (o *Outputter) Infof(format string, a ...interface{}) {
	o.Info(fmt.Sprintf(format, a))
}

func (o *Outputter) Debug(a ...interface{}) {
	if o.logfile != nil {
		fmt.Fprint(o.logfile, a)
	}
	if o.verbose {
		fmt.Fprint(o.outw, a)
	}
}

func (o *Outputter) Debugf(format string, a ...interface{}) {
	o.Debug(fmt.Sprintf(format, a))
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
