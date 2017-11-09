package command

import (
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//Job command
type Job struct {
	*command
}

//JobFactory returns a factory method for the join command
func JobFactory() (cli.Command, error) {
	comm, err := newCommand("nerd job <subcommand>", "Manage the lifecycle of compute jobs.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &Job{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//HelpTemplate provides a template for the help command, which excludes the "failure", "heartbeat", "receive", and "success" subcommands
func (cmd *Job) HelpTemplate() string {
	return fmt.Sprintf(`{{.Help}}{{if gt (len .Subcommands) 0}}
Subcommands:
{{- range $value := .Subcommands }}{{if and (and (ne "%v" $value.Name) (ne "%v" $value.Name)) (and (ne "%v" $value.Name) (ne "%v" $value.Name))}}
    {{ $value.NameAligned }}    {{ $value.Synopsis }}{{ end }}{{ end }}
{{- end }}

`, "failure", "heartbeat", "receive", "success")
}

//DoRun is called by run and allows an error to be returned
func (cmd *Job) DoRun(args []string) (err error) {
	return errShowHelp("Not enough arguments, see below for usage.")
}
