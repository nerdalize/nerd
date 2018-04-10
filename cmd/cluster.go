package cmd

import (
	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//Cluster command
type Cluster struct {
	*command
}

//ClusterFactory creates the command
func ClusterFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &Cluster{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, nil, flags.None, "nerd cluster")

	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *Cluster) Execute(args []string) (err error) { return errShowHelp("") }

// Description returns long-form help text
func (cmd *Cluster) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *Cluster) Synopsis() string {
	return "Group of commands used to manage clusters."
}

// Usage shows usage
func (cmd *Cluster) Usage() string { return "nerd cluster <subcommand>" }
