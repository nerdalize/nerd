package command

import (
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	"github.com/pkg/errors"
)

//DatasetList command
type DatasetList struct {
	*command
}

//DatasetListFactory returns a factory method for the join command
func DatasetListFactory() (cli.Command, error) {
	comm, err := newCommand("nerd dataset list", "Show a list of all datasets.", "", nil)
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

	projectID, err := ss.RequireProjectID()
	if err != nil {
		return HandleError(err)
	}

	datasets, err := bclient.ListDatasets(projectID)
	if err != nil {
		return HandleError(err)
	}

	header := "DATASET ID\tCREATED"
	pretty := "{{range $i, $x := $.Datasets}}{{$x.DatasetID}}\t{{$x.CreatedAt | fmtUnixAgo }}\n{{end}}"
	raw := "{{range $i, $x := $.Datasets}}{{$x.DatasetID}}\t{{$x.CreatedAt}}\n{{end}}"
	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(datasets, header, pretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(datasets, raw),
		format.OutputTypeJSON:   format.NewJSONDecorator(datasets.Datasets),
	})

	return nil
}
