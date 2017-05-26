package command

import (
	"os"

	"github.com/mitchellh/cli"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

// SecretList command
type SecretList struct {
	*command
}

// SecretListFactory returns a factory method for the join command
func SecretListFactory() (cli.Command, error) {
	comm, err := newCommand("nerd secret list", "show a list of all secrets in the current project", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &SecretList{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *SecretList) DoRun(args []string) (err error) {
	bclient, err := NewClient(cmd.ui, cmd.config, cmd.session)
	if err != nil {
		HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}
	out, err := bclient.ListSecrets(ss.Project.Name)
	if err != nil {
		HandleError(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ProjectID", "Name", "Key", "Value"})
	for _, t := range out.Secrets {
		row := []string{}
		row = append(row, t.ProjectID)
		row = append(row, t.Name)
		row = append(row, t.Key)
		row = append(row, t.Value)
		table.Append(row)
	}

	table.Render()
	return nil
}
