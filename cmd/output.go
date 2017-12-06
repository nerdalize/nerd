package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/mitchellh/cli"
	"github.com/sirupsen/logrus"
)

//Output standardizes the way in which we would like to capture output
//throughout the program
type Output struct {
	uiw *cli.UiWriter
	cli.Ui
}

//NewOutput sets up our standardized program outputter
func NewOutput(ui cli.Ui) *Output {
	return &Output{Ui: ui, uiw: &cli.UiWriter{Ui: ui}}
}

//Logger returns a logrus logger that writes to the UIs Stderr
func (o *Output) Logger(level logrus.Level) *logrus.Logger {
	logs := logrus.New()
	logs.Out = o.uiw
	logs.Level = level
	return logs
}

//Errorf prints a formatted error to ErrorOutput
func (o *Output) Errorf(format string, a ...interface{}) {
	o.Error(fmt.Sprintf(format, a...))
}

//Infof prints a formatted message to
func (o *Output) Infof(format string, a ...interface{}) {
	o.Error(fmt.Sprintf(format, a...))
}

//Table will print a table
func (o *Output) Table(header []string, rows [][]string) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	if len(header) > 0 {
		fmt.Fprintln(w, strings.Join(header, "\t")+"\t")
	}

	for _, r := range rows {
		fmt.Fprintln(w, strings.Join(r, "\t")+"\t")
	}

	return w.Flush()
}
