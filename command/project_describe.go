package command

import (
	"bytes"

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

	datasets, err := bclient.ListDatasets(project)
	if err != nil {
		return HandleError(err)
	}

	var tasks []*v1payload.TaskSummary
	for _, workload := range workloads.Workloads {
		tmp, err := bclient.ListTasks(ss.Project.Name, workload.WorkloadID, true)
		if err != nil {
			return HandleError(err)
		}
		tasks = append(tasks, tmp.Tasks...)
	}

	secrets, err := bclient.ListSecrets(project)
	if err != nil {
		return HandleError(err)
	}

	plans, err := bclient.ListPlans(project)
	if err != nil {
		return HandleError(err)
	}

	out := struct {
		Workloads []*v1payload.WorkloadSummary
		Tasks     []*v1payload.TaskSummary
		Datasets  []*v1payload.DatasetSummary
		Secrets   []*v1payload.SecretSummary
		Plans     []*v1payload.PlanSummary
	}{
		workloads.Workloads,
		tasks,
		datasets.Datasets,
		secrets.Secrets,
		plans.Plans,
	}

	var pretty bytes.Buffer
	pretty.WriteString("WORKLOAD ID\tIMAGE\tINPUT DATASET ID\tINSTANCES\tCREATED\n{{range $i, $x := $.Workloads}}{{$x.WorkloadID}}\t{{$x.Image}}\t{{$x.InputDatasetID}}\t{{$x.NrOfWorkers}}\t{{$x.CreatedAt | fmtUnixAgo }}\n{{end}}")
	pretty.WriteString("\nTASK ID\tWORKLOAD ID\tCMD\tOUTPUT DATASET ID\tSTATUS\tCREATED\n{{range $i, $x := $.Tasks}}{{$x.TaskID}}\t{{$x.WorkloadID}}\t{{$x.Cmd | testTruncate }}\t{{$x.OutputDatasetID}}\t{{$x.Status}}\t{{$x.TaskID | fmtUnixNanoAgo}}\n{{end}}")
	pretty.WriteString("\nDATASET ID\tCREATED\n{{range $i, $x := $.Datasets}}{{$x.DatasetID}}\t{{$x.CreatedAt | fmtUnixAgo }}\n{{end}}")
	pretty.WriteString("\nSECRET NAME\tTYPE\n{{range $i, $x := $.Secrets}}{{$x.Name}}\t{{$x.Type}}\n{{end}}")
	pretty.WriteString("\nPLAN ID\tCPU REQUEST\tMEMORY REQUEST\n{{range $i, $x := $.Plans}}{{$x.PlanID}}\t{{$x.RequestsCPU}}\t{{$x.RequestsMemory}}\n{{end}}")

	var raw bytes.Buffer
	raw.WriteString("WORKLOAD ID\tIMAGE\tINPUT DATASET ID\tINSTANCES\tCREATED\n{{range $i, $x := $.Workloads}}{{$x.WorkloadID}}\t{{$x.Image}}\t{{$x.InputDatasetID}}\t{{$x.NrOfWorkers}}\t{{$x.CreatedAt}}\n{{end}}")
	raw.WriteString("\nDATASET ID\tCREATED\n{{range $i, $x := $.Datasets}}{{$x.DatasetID}}\t{{$x.CreatedAt}}\n{{end}}")
	raw.WriteString("\nTASK ID\tWORKLOAD ID\tCMD\tOUTPUT DATASET ID\tSTATUS\tCREATED\n{{range $i, $x := $.Tasks}}{{$x.TaskID}}\t{{$x.WorkloadID}}\t{{$x.Cmd}}\t{{$x.OutputDatasetID}}\t{{$x.Status}}\t{{$x.TaskID}}\n{{end}}")
	raw.WriteString("\nSECRET NAME\tTYPE\n{{range $i, $x := $.Secrets}}{{$x.Name}}\t{{$x.Type}}\n{{end}}")
	raw.WriteString("\nPLAN ID\tCPU REQUEST\tMEMORY REQUEST\n{{range $i, $x := $.Plans}}{{$x.PlanID}}\t{{$x.RequestsCPU}}\t{{$x.RequestsMemory}}\n{{end}}")

	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(out, "Project details", pretty.String()),
		format.OutputTypeRaw:    format.NewTmplDecorator(out, raw.String()),
		format.OutputTypeJSON:   format.NewJSONDecorator(out),
	})

	return nil
}
