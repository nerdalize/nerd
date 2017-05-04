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
	version = nerd.BuiltFromSourceVersion
	commit  = "0000000"
)

func init() {
	opts := new(command.ConfOpts)
	_, err := flags.NewParser(opts, flags.None).ParseArgs(os.Args[1:])
	if err == nil {
		conf.SetLocation(opts.ConfigFile)
		nerd.SetupLogging(opts.VerboseOutput, opts.JSONOutput)
		nerd.VersionMessage(version)
	}
}

func main() {
	c := cli.NewCLI(name, fmt.Sprintf("%s (%s)", version, commit))
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"login":            command.LoginFactory,
		"worker":           command.WorkerFactory,
		"worker start":     command.WorkerStartFactory,
		"worker stop":      command.WorkerStopFactory,
		"worker work":      command.WorkerWorkFactory,
		"dataset":          command.DatasetFactory,
		"dataset upload":   command.DatasetUploadFactory,
		"dataset download": command.DatasetDownloadFactory,
		"queue":            command.QueueFactory,
		"queue describe":   command.QueueDescribeFactory,
		"queue create":     command.QueueCreateFactory,
		"queue delete":     command.QueueDeleteFactory,
		"task":             command.TaskFactory,
		"task list":        command.TaskListFactory,
		"task start":       command.TaskStartFactory,
		"task stop":        command.TaskStopFactory,
		"task describe":    command.TaskDescribeFactory,
		"task receive":     command.TaskReceiveFactory,
		"task heartbeat":   command.TaskHeartbeatFactory,
		"task success":     command.TaskSuccessFactory,
		"task failure":     command.TaskFailureFactory,
	}

	status, err := c.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s", name, err)
	}

	os.Exit(status)
}
