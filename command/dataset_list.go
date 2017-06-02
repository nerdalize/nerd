package command

import (
	"context"
	"os"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/aws"
	v1datatransfer "github.com/nerdalize/nerd/nerd/service/datatransfer/v1"
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
	bclient, err := NewClient(cmd.config, cmd.session)
	if err != nil {
		HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}
	dataOps, err := aws.NewDataClient(
		aws.NewNerdalizeCredentials(bclient, ss.Project.Name),
		ss.Project.AWSRegion,
	)
	if err != nil {
		HandleError(errors.Wrap(err, "could not create aws dataops client"))
	}
	out, err := bclient.ListDatasets(ss.Project.Name)
	if err != nil {
		HandleError(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ProjectID", "DatasetID", "Created", "Size"})
	for _, t := range out.Datasets {
		var size int64
		size, err = v1datatransfer.GetRemoteDatasetSize(context.Background(), bclient, dataOps, ss.Project.Name, t.DatasetID)
		if err != nil {
			HandleError(err)
		}
		row := []string{}
		row = append(row, t.ProjectID)
		row = append(row, t.DatasetID)
		row = append(row, humanize.Time(time.Unix(t.CreatedAt, 0)))
		row = append(row, humanize.Bytes(uint64(size)))
		table.Append(row)
	}

	table.Render()
	return nil
}
