package command

import (
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	"github.com/pkg/errors"
)

//TaskList command
type TaskList struct {
	*command
}

//TaskListFactory returns a factory method for the join command
func TaskListFactory() (cli.Command, error) {
	comm, err := newCommand("nerd task list <workload-id>", "show a list of all task currently in a queue", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &TaskList{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *TaskList) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}
	out, err := bclient.ListTasks(ss.Project.Name, args[0], false)
	if err != nil {
		return HandleError(err)
	}

	header := "TaskID\tCmd\tOutput\tStatus"
	pretty := "{{range $i, $x := $.Tasks}}{{$x.TaskID}}\t{{$x.Cmd}}\t{{$x.OutputDatasetID}}\t{{$x.Status}}\n{{end}}"
	raw := "{{range $i, $x := $.Tasks}}{{$x.TaskID}}\t{{$x.Cmd}}\t{{$x.OutputDatasetID}}\t{{$x.Status}}\n{{end}}"
	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(out, header, pretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(out, raw),
		format.OutputTypeJSON:   format.NewJSONDecorator(out.Tasks),
	})

	return nil
}
