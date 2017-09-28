package command

import (
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	"github.com/pkg/errors"
)

//PlanDescribe command
type PlanDescribe struct {
	*command
}

//PlanDescribeFactory returns a factory method for the plan describe command
func PlanDescribeFactory() (cli.Command, error) {
	comm, err := newCommand("nerd plan describe <name>", "Show more information about a specific plan.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &PlanDescribe{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *PlanDescribe) DoRun(args []string) (err error) {
	var (
		tmplPretty, tmplRaw string
	)

	if len(args) < 1 {
		return errShowHelp("Not enough arguments, see below for usage.")
	}

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

	out, err := bclient.DescribePlan(ss.Project.Name, args[0])
	if err != nil {
		return HandleError(err)
	}

	// Add used and remaining credits
	tmplPretty = `Name:			{{.PlanID}}
Project:		{{.ProjectID}}
CPU:			{{.RequestsCPU}}
Memory:			{{.RequestsMemory}}
Used CPU:		{{.RequestsCPU}}
Used Memory:	{{.RequestsMemory}}
`

	tmplRaw = `ID:		{{.PlanID}}
Project:	{{.ProjectID}}
CPU:		{{.RequestsCPU}}
Memory:		{{.RequestsMemory}}
Used CPU:		{{.RequestsCPU}}
Used Memory:	{{.RequestsMemory}}
`

	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(out, "Plan Details:", tmplPretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(out, tmplRaw),
		format.OutputTypeJSON:   format.NewJSONDecorator(out),
	})

	return nil
}
