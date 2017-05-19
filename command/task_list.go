package command

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/olekukonko/tablewriter"
)

//TaskListOpts describes command options
type TaskListOpts struct {
	NerdOpts
}

//TaskList command
type TaskList struct {
	*command
	opts   *TaskListOpts
	parser *flags.Parser
}

//TaskListFactory returns a factory method for the join command
func TaskListFactory() (cli.Command, error) {
	cmd := &TaskList{
		command: &command{
			help:     "",
			synopsis: "show a list of all task currently in a queue",
			parser:   flags.NewNamedParser("nerd task list <queue-id>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &TaskListOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskList) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	config, err := conf.Read()
	if err != nil {
		HandleError(err)
	}

	bclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err)
	}

	out, err := bclient.ListTasks(config.CurrentProject.Name, args[0])
	if err != nil {
		HandleError(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"QueueID", "TaskID", "Status"})
	for _, t := range out.Tasks {
		row := []string{}
		row = append(row, t.QueueID)
		row = append(row, fmt.Sprintf("%d", t.TaskID))
		row = append(row, t.Status)
		table.Append(row)
	}

	table.Render()
	return nil
}
