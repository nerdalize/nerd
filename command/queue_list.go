package command

import (
	"os"

	"github.com/mitchellh/cli"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

//QueueList command
type QueueList struct {
	*command
}

//QueueListFactory returns a factory method for the join command
func QueueListFactory() (cli.Command, error) {
	comm, err := newCommand("nerd queue list", "show a list of all queues in the current project", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &QueueList{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *QueueList) DoRun(args []string) (err error) {
	bclient, err := NewClient(cmd.ui, cmd.config, cmd.session)
	if err != nil {
		HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}
	out, err := bclient.ListQueues(ss.Project.Name)
	if err != nil {
		HandleError(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ProjectID", "QueueID"})
	for _, t := range out.Queues {
		row := []string{}
		row = append(row, t.ProjectID)
		row = append(row, t.QueueID)
		table.Append(row)
	}

	table.Render()
	return nil
}
