package command

import (
	"os"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/mitchellh/cli"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

//DatasetList command
type DatasetList struct {
	*command
}

//DatasetListFactory returns a factory method for the join command
func DatasetListFactory() (cli.Command, error) {
	comm, err := newCommand("nerd dataset list", "show a list of all datasets", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &DatasetList{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *DatasetList) DoRun(args []string) (err error) {
	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}
	out, err := bclient.ListDatasets(ss.Project.Name)
	if err != nil {
		return HandleError(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ProjectID", "DatasetID", "Created"})
	for _, t := range out.Datasets {
		row := []string{}
		row = append(row, t.ProjectID)
		row = append(row, t.DatasetID)
		row = append(row, humanize.Time(time.Unix(t.CreatedAt, 0)))
		table.Append(row)
	}

	table.Render()
	return nil
}
