package main

import (
	"fmt"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/evalphobia/logrus_sentry"
	"github.com/nerdalize/nerd/command"
	"github.com/rifflock/lfshook"

	"github.com/mitchellh/cli"
	homedir "github.com/mitchellh/go-homedir"
)

var (
	name    = "nerd"
	version = "build.from.src"
	commit  = "0000000"
)

func init() {
	logrus.SetLevel(logrus.WarnLevel)
	f, err := homedir.Expand("~/.nerd/log")
	if err == nil {
		logrus.AddHook(lfshook.NewHook(lfshook.PathMap{
			logrus.InfoLevel:  f,
			logrus.WarnLevel:  f,
			logrus.ErrorLevel: f,
		}))
	}
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

func main() {
	c := cli.NewCLI(name, fmt.Sprintf("%s (%s)", version, commit))
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"login":    command.LoginFactory(),
		"upload":   command.UploadFactory(),
		"run":      command.RunFactory(),
		"logs":     command.LogsFactory(),
		"work":     command.WorkFactory(),
		"status":   command.StatusFactory(),
		"download": command.DownloadFactory(),
	}

	status, err := c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s", name, err)
	}

	os.Exit(status)
}
