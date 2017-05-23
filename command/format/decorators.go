package format

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"text/tabwriter"

	"github.com/pkg/errors"
)

type jsonDecorator struct {
	v interface{}
}

func JSONDecorator(v interface{}) *jsonDecorator {
	return &jsonDecorator{
		v: v,
	}
}

func (d *jsonDecorator) Decorate(out io.Writer) error {
	enc := json.NewEncoder(out)
	return enc.Encode(d.v)
}

type tmplDecorator struct {
	v    interface{}
	tmpl string
}

func TmplDecorator(v interface{}, tmpl string) *tmplDecorator {
	return &tmplDecorator{
		v:    v,
		tmpl: tmpl,
	}
}
func (d *tmplDecorator) Decorate(out io.Writer) error {
	tmpl, err := template.New("tmpl").Parse(d.tmpl)
	if err != nil {
		return errors.Wrapf(err, "failed to create new output template for template %v", d.tmpl)
	}
	err = tmpl.Execute(out, d.v)
	if err != nil {
		return errors.Wrap(err, "failed to parse output into template")
	}
	return nil
}

type tableDecorator struct {
	v      interface{}
	header string
	tmpl   string
}

func TableDecorator(v interface{}, header, tmpl string) *tableDecorator {
	return &tableDecorator{
		v:      v,
		header: header,
		tmpl:   tmpl,
	}
}
func (d *tableDecorator) Decorate(out io.Writer) error {
	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, d.header)
	tmpl, err := template.New("tmpl").Parse(d.tmpl)
	if err != nil {
		return errors.Wrapf(err, "failed to create new output template for template %v", d.tmpl)
	}
	err = tmpl.Execute(w, d.v)
	if err != nil {
		return errors.Wrap(err, "failed to parse output into template")
	}
	w.Flush()
	return nil
}

type notImplDecorator struct {
	outputType OutputType
}

func NotImplDecorator(ot OutputType) *notImplDecorator {
	return &notImplDecorator{
		outputType: ot,
	}
}
func (d *notImplDecorator) Decorate(out io.Writer) error {
	fmt.Fprintf(out, "the selected output format (%v) is not supported", d.outputType)
	return nil
}
