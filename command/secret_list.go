package command

import (
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	"github.com/pkg/errors"
)

// SecretListOpts describes the options to the SecretList command
type SecretListOpts struct {
	Type string `long:"type" default:"all" default-mask:"" description:"Type of secret to display, defaults to all."`
}

//SecretList command
type SecretList struct {
	*command
	opts *SecretListOpts
}

// SecretListFactory returns a factory method for the secret list command
func SecretListFactory() (cli.Command, error) {
	opts := &SecretListOpts{}
	comm, err := newCommand("nerd secret list", "Show a list of all secrets in the current project.", "", opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &SecretList{
		command: comm,
		opts:    opts,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *SecretList) DoRun(args []string) (err error) {
	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}
	out, err := bclient.ListSecrets(ss.Project.Name)
	if err != nil {
		return HandleError(err)
	}

	header := "Secret name\tType"
	pretty := "{{range $i, $x := $.Secrets}}{{$x.Name}}\t{{$x.Type}}\n{{end}}"
	raw := "{{range $i, $x := $.Secrets}}{{$x.Name}}\t{{$x.Type}}\n{{end}}"

	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(out, header, pretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(out, raw),
		format.OutputTypeJSON:   format.NewJSONDecorator(out.Secrets),
	})

	return nil
}
