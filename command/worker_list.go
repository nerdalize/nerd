package command

import (
	"os"

	"github.com/mitchellh/cli"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

//WorkerList command
type WorkerList struct {
	*command
}

//WorkerListFactory returns a factory method for the join command
func WorkerListFactory() (cli.Command, error) {
	comm, err := newCommand("nerd worker list", "show a list of all workers in the current project", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &WorkerList{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkerList) DoRun(args []string) (err error) {
	bclient, err := NewClient(cmd.ui, cmd.config, cmd.session)
	if err != nil {
		HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}

	out, err := bclient.ListWorkers(ss.Project.Name)
	if err != nil {
		HandleError(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ProjectID", "WorkerID"})
	for _, t := range out.Workers {
		row := []string{}
		row = append(row, t.ProjectID)
		row = append(row, t.WorkerID)
		table.Append(row)
	}

	table.Render()
	return nil
}
