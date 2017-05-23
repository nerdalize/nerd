package command

import (
	"fmt"
	"os"

	"github.com/mitchellh/cli"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

//TaskList command
type TaskList struct {
	*command
}

//TaskListFactory returns a factory method for the join command
func TaskListFactory() (cli.Command, error) {
	comm, err := newCommand("nerd task list <queue-id>", "show a list of all task currently in a queue", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &TaskList{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskList) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	bclient, err := NewClient(cmd.ui, cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}
	out, err := bclient.ListTasks(ss.Project.Name, args[0])
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
