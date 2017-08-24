package command

import (
	"os"

	"github.com/mitchellh/cli"
	"github.com/olekukonko/tablewriter"
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
		return errors.Wrap(errShowHelp("show help"), "Not enough arguments, see below for usage.")
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

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Type", "Key", "Value", "Username", "Password"})
	row := []string{}
	row = append(row, out.Name)
	row = append(row, out.Type)
	row = append(row, out.Key)
	row = append(row, out.Value)
	row = append(row, out.DockerUsername)
	row = append(row, out.DockerPassword)
	table.Append(row)

	table.Render()

	return nil
}
