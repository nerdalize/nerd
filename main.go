package main

import (
	"fmt"
	"os"

	"github.com/nerdalize/nerd/cmd"
	"github.com/nerdalize/nerd/command"
	"github.com/nerdalize/nerd/nerd"

	"github.com/mitchellh/cli"
)

var (
	name    = "nerd"
	version = nerd.BuiltFromSourceVersion
	commit  = "0000000"
)

func create() *cli.CLI {
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "-V" || arg == "-version" || arg == "--version" {
			newArgs := make([]string, len(args)+1)
			newArgs[0] = "version"
			copy(newArgs[1:], args)
			args = newArgs //overwrite args to use the version subcommand
			break
		}
	}

	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	c := &cli.CLI{
		Name:           name,
		Args:           args,
		HiddenCommands: []string{},
		Commands: map[string]cli.CommandFactory{
			"version":          command.CreateVersionFactory(version, commit),
			"login":            command.LoginFactory,
			"dataset":          cmd.DatasetFactory(ui),
			"dataset upload":   cmd.DatasetUploadFactory(ui),
			"dataset download": cmd.DatasetDownloadFactory(ui),
			"dataset list":     cmd.DatasetListFactory(ui),
			"dataset delete":   cmd.DatasetDeleteFactory(ui),
			"job":              cmd.JobFactory(ui),
			"job run":          cmd.JobRunFactory(ui),
			"job delete":       cmd.JobDeleteFactory(ui),
			"job list":         cmd.JobListFactory(ui),
			"job logs":         cmd.JobLogsFactory(ui),
		},
	}

	return c
}

func main() {
	status, err := create().Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s", name, err)
	}

	os.Exit(status)
}
