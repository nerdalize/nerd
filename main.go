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
		Name: name,
		Args: args,
		HiddenCommands: []string{
			"job",
			"project expel",
			"project place",
			"task receive",
			"task heartbeat",
			"task success",
			"task failure",
		},
		Commands: map[string]cli.CommandFactory{
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
			"job":               cmd.JobFactory(ui),
			"job run":           cmd.JobRunFactory(ui),
			"job delete":        cmd.JobDeleteFactory(ui),
			"job list":          cmd.JobListFactory(ui),
			"task":              command.TaskFactory,
			"task list":         command.TaskListFactory,
			"task create":       command.TaskCreateFactory,
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
