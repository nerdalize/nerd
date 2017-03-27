package main

import (
	"fmt"
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/nerdalize/nerd/command"
	"github.com/nerdalize/nerd/nerd"
	"github.com/nerdalize/nerd/nerd/conf"

	"github.com/mitchellh/cli"
)

var (
	name    = "nerd"
	version = "build.from.src"
	commit  = "0000000"
)

type ConfOpts struct {
	ConfigFile string `long:"config" default:"" default-mask:"" env:"CONFIG" description:"location of config file"`
}

func init() {
	opts := new(ConfOpts)
	_, err := flags.NewParser(opts, flags.IgnoreUnknown).ParseArgs(os.Args[1:])
	if err == nil {
		conf.SetLocation(opts.ConfigFile)
	}
	nerd.SetupLogging()
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
