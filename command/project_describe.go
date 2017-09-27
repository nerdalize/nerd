package command

import (
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
	"github.com/pkg/errors"
)

//ProjectDescribe command
type ProjectDescribe struct {
	*command
}

//ProjectDescribeFactory returns a factory method for the project describe command
func ProjectDescribeFactory() (cli.Command, error) {
	comm, err := newCommand("nerd project describe <projectID>", "Show more information about a specific project.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &ProjectDescribe{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *ProjectDescribe) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return errShowHelp("Not enough arguments, see below for usage.")
	}
	project := args[0]

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}
	_, err = ss.RequireProjectID()
	if err != nil {
		return HandleError(err)
	}

	workloads, err := bclient.ListWorkloads(project)
	if err != nil {
		return HandleError(err)
	}
	header := "WORKLOAD ID\tIMAGE\tINPUT\tCREATED AT"
	pretty := "{{range $i, $x := $.Workloads}}{{$x.WorkloadID}}\t{{$x.Image}}\t{{$x.InputDatasetID}}\t{{$x.CreatedAt | fmtUnixAgo }}\n{{end}}"
	raw := "{{range $i, $x := $.Workloads}}{{$x.WorkloadID}}\t{{$x.Image}}\t{{$x.InputDatasetID}}\t{{$x.CreatedAt}}\n{{end}}"
	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(workloads, header, pretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(workloads, raw),
		format.OutputTypeJSON:   format.NewJSONDecorator(workloads.Workloads),
	})

	datasets, err := bclient.ListDatasets(project)
	if err != nil {
		return HandleError(err)
	}
	header = "\nDATASET ID\tCREATED AT"
	pretty = "{{range $i, $x := $.Datasets}}{{$x.DatasetID}}\t{{$x.CreatedAt | fmtUnixAgo }}\n{{end}}"
	raw = "{{range $i, $x := $.Datasets}}{{$x.DatasetID}}\t{{$x.CreatedAt}}\n{{end}}"
	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(datasets, header, pretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(datasets, raw),
		format.OutputTypeJSON:   format.NewJSONDecorator(datasets.Datasets),
	})

	var tasks []*v1payload.TaskSummary
	for _, workload := range workloads.Workloads {
		tmp, err := bclient.ListTasks(ss.Project.Name, workload.WorkloadID, true)
		if err != nil {
			return HandleError(err)
		}
		tasks = append(tasks, tmp.Tasks...)
	}

	header = "\nTASK ID\tWORKLOAD ID\tCMD\tOUTPUT ID\tSTATUS\tCREATED AT"
	pretty = "{{range $i, $x := $}}{{$x.TaskID}}\t{{$x.WorkloadID}}\t{{$x.Cmd}}\t{{$x.OutputDatasetID}}\t{{$x.Status}}\t{{$x.TaskID | fmtUnixNanoAgo}}\n{{end}}"
	raw = "{{range $i, $x := $}}{{$x.TaskID}}\t{{$x.WorkloadID}}\t{{$x.Cmd}}\t{{$x.OutputDatasetID}}\t{{$x.Status}}\t{{$x.TaskID}}\n{{end}}"
	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(tasks, header, pretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(tasks, raw),
		format.OutputTypeJSON:   format.NewJSONDecorator(tasks),
	})

	secrets, err := bclient.ListSecrets(project)
	if err != nil {
		return HandleError(err)
	}
	header = "\nSECRET NAME\tTYPE"
	pretty = "{{range $i, $x := $.Secrets}}{{$x.Name}}\t{{$x.Type}}\n{{end}}"
	raw = "{{range $i, $x := $.Secrets}}{{$x.Name}}\t{{$x.Type}}\n{{end}}"
	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(secrets, header, pretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(secrets, raw),
		format.OutputTypeJSON:   format.NewJSONDecorator(secrets.Secrets),
	})

	plans, err := bclient.ListPlans(project)
	if err != nil {
		return HandleError(err)
	}
	header = "\nPLANS"
	pretty = "{{range $i, $x := $.Plans}}{{$x.UID}}\n{{end}}"
	raw = "{{range $i, $x := $.Plans}}{{$x.UID}}\t{{$x.CPU}}\n{{end}}"
	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(plans, header, pretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(plans, raw),
		format.OutputTypeJSON:   format.NewJSONDecorator(plans.Plans),
	})
	return nil
}
