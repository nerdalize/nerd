package main

import (
	"fmt"
	"os"

	"github.com/nerdalize/nerd/command"
	"github.com/nerdalize/nerd/nerd"

	"github.com/mitchellh/cli"
)

var (
	name    = "nerd"
	version = nerd.BuiltFromSourceVersion
	commit  = "0000000"
)

func init() {
	nerd.SetupLogging()
	nerd.VersionMessage(version)
}

func main() {
	c := cli.NewCLI(name, fmt.Sprintf("%s (%s)", version, commit))
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"login":            command.LoginFactory,
		"worker":           command.WorkerFactory,
		"worker start":     command.WorkerStartFactory,
		"worker stop":      command.WorkerStopFactory,
		"worker list":      command.WorkerListFactory,
		"worker work":      command.WorkerWorkFactory,
		"dataset":          command.DatasetFactory,
		"dataset upload":   command.DatasetUploadFactory,
		"dataset download": command.DatasetDownloadFactory,
		"queue":            command.QueueFactory,
		"queue list":       command.QueueListFactory,
		"queue describe":   command.QueueDescribeFactory,
		"queue create":     command.QueueCreateFactory,
		"queue delete":     command.QueueDeleteFactory,
		"project":          command.ProjectFactory,
		"project place":    command.ProjectPlaceFactory,
		"project expel":    command.ProjectExpelFactory,
		"project set":      command.ProjectSetFactory,
		"project list":     command.ProjectListFactory,
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
