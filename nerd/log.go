package nerd

import (
	"fmt"
	"io"
	"os"

	"github.com/Sirupsen/logrus"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/nerdalize/nerd/nerd/conf"
)

//PlainFormatter is a Logrus formatter that only includes the log message.
type PlainFormatter struct {
}

//Format removes all formatting for the PlainFormatter.
func (f *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	msg := fmt.Sprintf("%s\n", entry.Message)
	return []byte(msg), nil
}

//SetupLogging sets up logrus.
func SetupLogging() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(new(PlainFormatter))
	addFSHook()
}

//addFSHook adds a filesystem logging hook to Logrus
func addFSHook() {
	c, err := conf.Read()
	if err == nil && c.EnableLogging {
		filename, err := homedir.Expand("~/.nerd/log")
		if err != nil {
			return
		}
		f, err := os.Open(filename)
		defer f.Close()
		if err != nil && os.IsNotExist(err) {
			f, err = os.Create(filename)
			defer f.Close()
			if err != nil {
				return
			}
		}
		if err != nil {
			return
		}
		logrus.SetOutput(io.MultiWriter(os.Stdout, f))
	}
}
