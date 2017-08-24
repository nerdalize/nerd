package command

import (
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//Task command
type Task struct {
	*command
}

//TaskFactory returns a factory method for the join command
func TaskFactory() (cli.Command, error) {
	comm, err := newCommand("nerd task <subcommand>", "Manage the lifecycle of compute tasks.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &Task{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//HelpTemplate provides a template for the help command, which excludes the "failure", "heartbeat", "receive", and "success" subcommands
func (cmd *Task) HelpTemplate() string {
	return fmt.Sprintf(`{{.Help}}{{if gt (len .Subcommands) 0}}
Subcommands:
{{- range $value := .Subcommands }}{{if and (and (ne "%v" $value.Name) (ne "%v" $value.Name)) (and (ne "%v" $value.Name) (ne "%v" $value.Name))}}
    {{ $value.NameAligned }}    {{ $value.Synopsis }}{{ end }}{{ end }}
{{- end }}

`, "failure", "heartbeat", "receive", "success")
}

//DoRun is called by run and allows an error to be returned
func (cmd *Task) DoRun(args []string) (err error) {
	return errShowHelp("show error")
}
