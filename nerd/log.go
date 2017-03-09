package nerd

import (
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/evalphobia/logrus_sentry"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/rifflock/lfshook"
)

//PlainFormatter is a Logrus formatter that only includes the log message.
type PlainFormatter struct {
}

func (f *PlainFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	msg := fmt.Sprintf("%s\n", entry.Message)
	return []byte(msg), nil
}

func SetupLogging() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(new(PlainFormatter))
	addFSHook()
	addSentryHook()
}

//addFSHook adds a filesystem logging hook to Logrus
func addFSHook() {
	c, err := conf.Read()
	if err == nil && c.EnableLogging {
		f, err := homedir.Expand("~/.nerd/log")
		if err == nil {
			logrus.AddHook(lfshook.NewHook(lfshook.PathMap{
				logrus.DebugLevel: f,
				logrus.InfoLevel:  f,
				logrus.WarnLevel:  f,
				logrus.ErrorLevel: f,
			}))
		}
	}
}

//addSentryHook adds a remote logging hook (Sentry.io) to Logrus
func addSentryHook() {
	user := os.Getenv("SENTRY_USER")
	pass := os.Getenv("SENTRY_PASS")
	if user != "" && pass != "" {
		// TODO: Add tags such as JWT and user ID
		dsn := fmt.Sprintf("https://%s:%s@sentry.io/144116", user, pass)
		hook, err := logrus_sentry.NewSentryHook(dsn, []logrus.Level{
			logrus.InfoLevel,
			logrus.WarnLevel,
		})
		if err == nil {
			hook.Timeout = time.Second
			logrus.AddHook(hook)
		}
	}
}
