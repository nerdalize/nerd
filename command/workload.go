package command

import (
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//Workload command
type Workload struct {
	*command
}

//WorkloadFactory returns a factory method for the join command
func WorkloadFactory() (cli.Command, error) {
	comm, err := newCommand("nerd workload <subcommand>", "Control compute capacity for working on tasks.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &Workload{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//HelpTemplate provides a template for the help command, which excludes the "workload work" command
func (cmd *Workload) HelpTemplate() string {
	return `{{.Help}}{{if gt (len .Subcommands) 0}}
Subcommands:
{{- range $value := .Subcommands }}{{if ne "work" $value.Name}}
    {{ $value.NameAligned }}    {{ $value.Synopsis }}{{ end }}{{ end }}
{{- end }}

`
}

//DoRun is called by run and allows an error to be returned
func (cmd *Workload) DoRun(args []string) (err error) {
	return errShowHelp
}
