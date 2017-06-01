package command

import (
	"os"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/mitchellh/cli"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

//WorkloadList command
type WorkloadList struct {
	*command
}

//WorkloadListFactory returns a factory method for the join command
func WorkloadListFactory() (cli.Command, error) {
	comm, err := newCommand("nerd workload list", "show a list of all workloads in the current project", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &WorkloadList{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkloadList) DoRun(args []string) (err error) {
	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}

	out, err := bclient.ListWorkloads(ss.Project.Name)
	if err != nil {
		return HandleError(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ProjectID", "WorkloadID", "Image", "Created"})
	for _, t := range out.Workloads {
		row := []string{}
		row = append(row, t.ProjectID)
		row = append(row, t.WorkloadID)
		row = append(row, t.Image)
		row = append(row, humanize.Time(time.Unix(t.CreatedAt, 0)))
		table.Append(row)
	}

	table.Render()
	return nil
}
