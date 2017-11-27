package commandv2

import "github.com/mitchellh/cli"

//JobRun command
type JobRun struct {
	*command
}

//JobRunFactory creates the command
func JobRunFactory() cli.CommandFactory {
	cmd := &JobRun{}
	cmd.command = createCommand(cmd.Execute, cmd.Description, cmd.Usage)
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *JobRun) Execute(args []string) (err error) { return nil }

// Description returns long-form help text
func (cmd *JobRun) Description() string { return PlaceholderHelp }

// Synopsis returns a one-line
func (cmd *JobRun) Synopsis() string { return PlaceholderSynopsis }

// Usage shows usage
func (cmd *JobRun) Usage() string { return PlaceholderUsage }
