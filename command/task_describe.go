package command

import (
	"strconv"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	"github.com/pkg/errors"
)

//TaskDescribe command
type TaskDescribe struct {
	*command
}

//TaskDescribeFactory returns a factory method for the join command
func TaskDescribeFactory() (cli.Command, error) {
	comm, err := newCommand("nerd task describe <workload-id> <task-id>", "return more information about a specific task", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &TaskDescribe{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskDescribe) DoRun(args []string) (err error) {
	if len(args) < 2 {
		return errors.Wrap(errShowHelp, "Not enough arguments, see below for usage.")
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	taskID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return HandleError(errors.Wrap(err, "invalid task ID, must be a number"))
	}

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}

	projectID, err := ss.RequireProjectID()
	if err != nil {
		return HandleError(err)
	}

	out, err := bclient.DescribeTask(projectID, args[0], taskID)
	if err != nil {
		return HandleError(err)
	}

	tmplPretty := `ID:	{{.TaskID}}
Cmd:	{{.Cmd}}
Output:	{{.OutputDatasetID}}
Status:	{{.Status}}
Created:	{{.TaskID | fmtUnixNanoAgo }}
`

	tmplRaw := `ID:	{{.TaskID}}
Cmd:	{{.Cmd}}
Output:	{{.OutputDatasetID}}
Status:	{{.Status}}
Created:	{{.TaskID}}
`

	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(out, "Workload Details:", tmplPretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(out, tmplRaw),
		format.OutputTypeJSON:   format.NewJSONDecorator(out),
	})

	return nil
}
