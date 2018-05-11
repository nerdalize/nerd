package cmd

import (
	"fmt"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd"
)

//Version command
type Version struct {
	version string
	commit  string
	*command
}

//VersionFactory returns a factory method for the join command
func VersionFactory(version, commit string, ui cli.Ui) cli.CommandFactory {

	cmd := &Version{
		version: version,
		commit:  commit,
	}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, nil, flags.None, "nerd version")
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute is called by run and allows an error to be returned
func (cmd *Version) Execute(args []string) (err error) {
	cmd.out.Info(fmt.Sprintf("%s (%s)", cmd.version, cmd.commit))
	nerd.VersionMessage(cmd.version)
	return nil
}

// Description returns long-form help text
func (cmd *Version) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *Version) Synopsis() string { return "Check the current version of the CLI" }

// Usage shows usage
func (cmd *Version) Usage() string { return "nerd version" }
