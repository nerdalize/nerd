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
	include = []string{
		"login",
		"version",
		"workload",
		"workload start",
		"workload stop",
		"workload list",
		"workload describe",
		"workload download",
		"worker",
		"worker logs",
		"dataset",
		"dataset upload",
		"dataset list",
		"dataset download",
		"project",
		"project set",
		"project list",
		"task",
		"task list",
		"task start",
		"task stop",
		"task describe",
		"secret",
		"secret list",
		"secret describe",
		"secret create",
		"secret delete",
	}
)

func create() *cli.CLI {

	//as seen in github.com/hashicorp/terraform/main.go
	args := os.Args[1:]
	for _, arg := range args {
		if arg == "-V" || arg == "-version" || arg == "--version" {
			newArgs := make([]string, len(args)+1)
			newArgs[0] = "version"
			copy(newArgs[1:], args)
			args = newArgs
			break
		}
	}

	//setup the cli
	c := cli.NewCLI(name, fmt.Sprintf("%s (%s)", version, commit))
	c.Args = args
	c.Commands = map[string]cli.CommandFactory{
		"version":           command.CreateVersionFactory(version, commit),
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
		"dataset download":  command.DatasetDownloadFactory,
		"dataset list":      command.DatasetListFactory,
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
		"secret":            command.SecretFactory,
		"secret list":       command.SecretListFactory,
		"secret describe":   command.SecretDescribeFactory,
		"secret create":     command.SecretCreateFactory,
		"secret delete":     command.SecretDeleteFactory,
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
