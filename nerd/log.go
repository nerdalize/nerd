package nerd

import (
	"fmt"

	"github.com/Sirupsen/logrus"
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
}
