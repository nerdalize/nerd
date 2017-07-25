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

// func init() {
// 	nerd.VersionMessage(version)
// }

func create() *cli.CLI {
	c := cli.NewCLI(name, fmt.Sprintf("%s (%s)", version, commit))
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"login":             command.LoginFactory,
		"workload":          command.WorkloadFactory,
		"workload start":    command.WorkloadStartFactory,
		"workload stop":     command.WorkloadStopFactory,
		"workload list":     command.WorkloadListFactory,
		"workload describe": command.WorkloadDescribeFactory,
		"workload download": command.WorkloadDownloadFactory,
		"workload work":     command.WorkloadWorkFactory,
		"worker":            command.WorkerFactory,
		"worker logs":       command.WorkerLogsFactory,
		"dataset":           command.DatasetFactory,
		"dataset upload":    command.DatasetUploadFactory,
		"dataset list":      command.DatasetListFactory,
		"dataset download":  command.DatasetDownloadFactory,
		"project":           command.ProjectFactory,
		"project place":     command.ProjectPlaceFactory,
		"project expel":     command.ProjectExpelFactory,
		"project set":       command.ProjectSetFactory,
		"project list":      command.ProjectListFactory,
		"task":              command.TaskFactory,
		"task list":         command.TaskListFactory,
		"task start":        command.TaskStartFactory,
		"task stop":         command.TaskStopFactory,
		"task describe":     command.TaskDescribeFactory,
		"task receive":      command.TaskReceiveFactory,
		"task heartbeat":    command.TaskHeartbeatFactory,
		"task success":      command.TaskSuccessFactory,
		"task failure":      command.TaskFailureFactory,
	}
	include := []string{
		"login",
		"workload",
		"workload start",
		"workload stop",
		"workload list",
		"workload describe",
		"worker",
		"worker logs",
		"dataset",
		"dataset upload",
		"dataset download",
		"project",
		"project set",
		"project list",
		"task",
		"task list",
		"task start",
		"task stop",
		"task describe",
	}
	c.HelpFunc = cli.FilteredHelpFunc(include, cli.BasicHelpFunc(name))
	return c
}

func main() {
	status, err := create().Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s", name, err)
	}

	os.Exit(status)
}
