package format

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"
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
	outputter := NewOutputter(outbuf, errbuf, log.New(errbuf, "", 0))
	outputter.SetOutputType("test")
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

func TestLogging(t *testing.T) {
	outbuf := bytes.NewBuffer(nil)
	errbuf := bytes.NewBuffer(nil)
	outputter := NewOutputter(outbuf, errbuf, log.New(errbuf, "", 0))
	// only show info messages
	t.Run("no debug", func(t *testing.T) {
		outputter.Logger.Println("[DEBUG] debugmsg")
		outputter.Logger.Println("[WARN] warnmsg")
		outputter.Logger.Println("[ERROR] errormsg")
		outputter.Logger.Println("[INFO] infomsg")
		outputter.Logger.Println("basemsg")
		exp := "[INFO] infomsg\nbasemsg\n"
		utils.Equals(t, exp, errbuf.String())
		utils.Equals(t, "", outbuf.String())
		errbuf.Reset()
	})
	// all messages should now show
	t.Run("debug", func(t *testing.T) {
		outputter.SetDebug(true)
		outputter.Logger.Println("[DEBUG] debugmsg")
		outputter.Logger.Println("[WARN] warnmsg")
		outputter.Logger.Println("[ERROR] errormsg")
		outputter.Logger.Println("[INFO] infomsg")
		outputter.Logger.Println("basemsg")
		s := errbuf.String()
		utils.Assert(t, strings.Contains(s, "[DEBUG] debugmsg"), "expected debug output", s)
		utils.Assert(t, strings.Contains(s, "[WARN] warnmsg"), "expected warn output", s)
		utils.Assert(t, strings.Contains(s, "[ERROR] errormsg"), "expected error output", s)
		utils.Assert(t, strings.Contains(s, "[INFO] infomsg"), "expected info output", s)
		utils.Assert(t, strings.Contains(s, "basemsg"), "expected base output", s)
		utils.Equals(t, "", outbuf.String())
		outputter.SetDebug(false)
		errbuf.Reset()
	})
	// everything should be logged even if debugging is false
	t.Run("logfile", func(t *testing.T) {
		path := tmpFile(t)
		outputter.SetLogToDisk(path)
		outputter.Logger.Println("[DEBUG] debugmsg")
		outputter.Logger.Println("[WARN] warnmsg")
		outputter.Logger.Println("[ERROR] errormsg")
		outputter.Logger.Println("[INFO] infomsg")
		outputter.Logger.Println("basemsg")
		outputter.Close()
		l, err := ioutil.ReadFile(path)
		utils.OK(t, err)
		s := string(l)
		utils.Assert(t, strings.Contains(s, "[DEBUG] debugmsg"), "expected debug output", s)
		utils.Assert(t, strings.Contains(s, "[WARN] warnmsg"), "expected warn output", s)
		utils.Assert(t, strings.Contains(s, "[ERROR] errormsg"), "expected error output", s)
		utils.Assert(t, strings.Contains(s, "[INFO] infomsg"), "expected info output", s)
		utils.Assert(t, strings.Contains(s, "basemsg"), "expected base output", s)
		utils.Equals(t, "", outbuf.String())
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
