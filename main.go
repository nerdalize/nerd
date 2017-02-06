package main

import (
	"fmt"
	"os"

	"github.com/nerdalize/nerd/command"

	"github.com/mitchellh/cli"
)

var (
	name    = "nerd"
	version = "build.from.src"
	commit  = "0000000"
)

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
