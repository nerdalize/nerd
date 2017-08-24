package command

import (
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	"github.com/pkg/errors"
)

//SecretDescribe command
type SecretDescribe struct {
	*command
}

//SecretDescribeFactory returns a factory method for the secret describe command
func SecretDescribeFactory() (cli.Command, error) {
	comm, err := newCommand("nerd secret describe <name>", "Show more information about a specific secret.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &SecretDescribe{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *SecretDescribe) DoRun(args []string) (err error) {
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
	out, err := bclient.DescribeSecret(ss.Project.Name, args[0])
	if err != nil {
		return HandleError(err)
	}

	tmplPretty := `Name:			{{.Name}}
	Type:			{{.Type}}
	Key:			{{.Key}}
	Value:			{{.Value}}
		`

	tmplRaw := `ID:			{{.Name}}
		Type:			{{.Type}}
		Key:			{{.Key}}
		Value:			{{.Value}}
		`

	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(out, "Secret Details:", tmplPretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(out, tmplRaw),
		format.OutputTypeJSON:   format.NewJSONDecorator(out),
	})

	return nil
}
