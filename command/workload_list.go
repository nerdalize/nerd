package command

import (
	"sort"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	"github.com/pkg/errors"
)

//WorkloadList command
type WorkloadList struct {
	*command
}

//WorkloadListFactory returns a factory method for the join command
func WorkloadListFactory() (cli.Command, error) {
	comm, err := newCommand("nerd workload list", "Show a list of all workloads in the current project.", "", nil)
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

	projectID, err := ss.RequireProjectID()
	if err != nil {
		return HandleError(err)
	}

	out, err := bclient.ListWorkloads(projectID)
	if err != nil {
		return HandleError(err)
	}

	sort.Slice(out.Workloads, func (i int, j int) bool {
		return out.Workloads[i].CreatedAt < out.Workloads[j].CreatedAt
	})

	header := "WORKLOAD ID\tIMAGE\tINPUT\tCREATED"
	pretty := "{{range $i, $x := $.Workloads}}{{$x.WorkloadID}}\t{{$x.Image}}\t{{$x.InputDatasetID}}\t{{$x.CreatedAt | fmtUnixAgo }}\n{{end}}"
	raw := "{{range $i, $x := $.Workloads}}{{$x.WorkloadID}}\t{{$x.Image}}\t{{$x.InputDatasetID}}\t{{$x.CreatedAt}}\n{{end}}"
	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(out, header, pretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(out, raw),
		format.OutputTypeJSON:   format.NewJSONDecorator(out.Workloads),
	})

	return nil
}
