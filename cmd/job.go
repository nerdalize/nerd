package cmd

import (
	"github.com/mitchellh/cli"
)

//Job command
type Job struct {
	*command
}

//JobFactory creates the command
func JobFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &Job{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd)

	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *Job) Execute(args []string) (err error) { return errShowHelp }

// Description returns long-form help text
func (cmd *Job) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *Job) Synopsis() string { return "Manage the lifecycle of compute jobs." }

// Usage shows usage
func (cmd *Job) Usage() string { return "nerd job <subcommand>" }
