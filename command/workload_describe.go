package command

import (
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	"github.com/pkg/errors"
)

//WorkloadDescribe command
type WorkloadDescribe struct {
	*command
}

//WorkloadDescribeFactory returns a factory method for the join command
func WorkloadDescribeFactory() (cli.Command, error) {
	comm, err := newCommand("nerd workload describe <workload-id>", "Return more information about a specific workload.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &WorkloadDescribe{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkloadDescribe) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return errors.Wrap(errShowHelp("show help"), "Not enough arguments, see below for usage.")
	}

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

	out, err := bclient.DescribeWorkload(projectID, args[0])
	if err != nil {
		return HandleError(err)
	}

	tmplPretty := `ID:			{{.WorkloadID}}
Image:			{{.Image}}
Input:			{{.InputDatasetID}}
Created:			{{.CreatedAt | fmtUnixAgo}}
Workers:			{{range $i, $x := $.Workers}}{{$x.WorkerID}} ({{$x.Status}}) {{end}}
	`

	tmplRaw := `ID:			{{.WorkloadID}}
	Image:			{{.Image}}
	Input:			{{.InputDatasetID}}
	Created:			{{.CreatedAt}}
	Workers:			{{range $i, $x := $.Workers}}{{$x.WorkerID}} ({{$x.Status}}) {{end}}
	`

	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(out, "Workload Details:", tmplPretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(out, tmplRaw),
		format.OutputTypeJSON:   format.NewJSONDecorator(out),
	})

	return nil
}
