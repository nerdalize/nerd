package format

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/nerdalize/nerd/nerd/utils"
)

type staticDecorator struct {
	contents string
}

func (d *staticDecorator) Decorate(out io.Writer) error {
	fmt.Fprintf(out, d.contents)
	return nil
}

func TestOutputter(t *testing.T) {
	outbuf := bytes.NewBuffer(nil)
	errbuf := bytes.NewBuffer(nil)
	outputter := &Outputter{
		outw:       outbuf,
		errw:       errbuf,
		outputType: "test",
	}
	exp := "a\nb\nc"
	t.Run("output", func(t *testing.T) {
		outputter.Output(DecMap{
			"test": &staticDecorator{exp},
		})
		utils.Equals(t, exp, outbuf.String())
		utils.Equals(t, "", errbuf.String())
	})
	t.Run("logging", func(t *testing.T) {
		path := tmpFile(t)
		outbuf = bytes.NewBuffer(nil)
		outputter.outw = outbuf
		outputter.SetLogToDisk(path)
		outputter.Output(DecMap{
			"test": &staticDecorator{exp},
		})
		err := outputter.Close()
		utils.OK(t, err)
		log, err := ioutil.ReadFile(path)
		utils.OK(t, err)
		utils.Equals(t, exp, string(log))
		utils.Equals(t, exp, outbuf.String())
		utils.Equals(t, "", errbuf.String())
	})
}

func tmpFile(t *testing.T) string {
	tmp, err := ioutil.TempFile("", "formatter_test")
	if err != nil {
		t.Fatalf("failed to setup temp file: %v", err)
	}
	defer tmp.Close()
	return tmp.Name()
}
