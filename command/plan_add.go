package command

import (
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
	"github.com/pkg/errors"
)

// PlanAddOpts describes the options to the PlanAdd command
// Don't know yet if they are required
type PlanAddOpts struct {
	CPU    string `long:"cpu" default:"1" default-mask:"" description:"CPU units you want to use in your on-demand plan."`
	Memory string `long:"memory" default:"1.5" default-mask:"" description:"Memory units you want to use in your on-demand plan."`
}

var (
	planAddUsage    = `nerd plan add <plan>`
	planAddSynopsis = "Add plan to your current project."
	planAddHelp     = `A plan reserved capacity for your project.`
)

//PlanAdd command
type PlanAdd struct {
	*command
	opts *PlanAddOpts
}

//PlanAddFactory returns a factory method for the plan add command
func PlanAddFactory() (cli.Command, error) {
	opts := &PlanAddOpts{}
	comm, err := newCommand(planAddUsage, planAddSynopsis, planAddHelp, opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add command")
	}
	cmd := &PlanAdd{
		command: comm,
		opts:    opts,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
// Get planID from args, check if it's an existing plan and get specs from auth service
// add it to universe
func (cmd *PlanAdd) DoRun(args []string) (err error) {
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

	planName := args[0]
	// check planName is a real plan -> auth_
	var out *v1payload.CreatePlanOutput
	out, err = bclient.CreatePlan(ss.Project.Name,
		planName,
		cmd.opts.CPU,
		cmd.opts.Memory,
	)
	if err != nil {
		return HandleError(err)
	}

	// TODO add type
	tmplPretty := `ID:	{{.PlanID}}
	CPU:	{{.RequestsCPU}}
	Memory:	{{.RequestsMemory}}
	`

	tmplRaw := `ID:	{{.PlanID}}
		CPU:	{{.RequestsCPU}}
		Memory:	{{.RequestsMemory}}
		`

	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(out, "Plan added:", tmplPretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(out, tmplRaw),
		format.OutputTypeJSON:   format.NewJSONDecorator(out),
	})

	return nil
}
