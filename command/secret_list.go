package command

import (
	"os"

	"github.com/mitchellh/cli"
	"github.com/olekukonko/tablewriter"
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

// SecretListFactory returns a factory method for the join command
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

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Type"})
	for _, t := range out.Secrets {
		row := []string{}
		row = append(row, t.Name)
		row = append(row, t.Type)
		table.Append(row)
	}

	table.Render()
	return nil
}
