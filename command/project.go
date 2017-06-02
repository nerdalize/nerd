package command

import (
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//Project command
type Project struct {
	*command
}

//ProjectFactory returns a factory method for the join command
func ProjectFactory() (cli.Command, error) {
	comm, err := newCommand("nerd project <subcommand>", "set and list projects", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &Project{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//HelpTemplate provides a template for the help command, which excludes the "expel" and "place" subcommands
func (cmd *Project) HelpTemplate() string {
	return fmt.Sprintf(`
{{.Help}}{{if gt (len .Subcommands) 0}}
Subcommands:
{{- range $value := .Subcommands }}{{if and (ne "%v" $value.Name) (ne "%v" $value.Name)}}
    {{ $value.NameAligned }}    {{ $value.Synopsis }}{{ end }}{{ end }}
{{- end }}
`, "place", "expel")
}

//DoRun is called by run and allows an error to be returned
func (cmd *Project) DoRun(args []string) (err error) {
	return errShowHelp
}
