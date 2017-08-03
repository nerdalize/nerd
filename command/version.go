package command

import (
	"fmt"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd"
	"github.com/pkg/errors"
)

//Version command
type Version struct {
	version string
	commit  string
	*command
}

//CreateVersionFactory returns a factory method for the join command
func CreateVersionFactory(version, commit string) cli.CommandFactory {
	return func() (cli.Command, error) {
		comm, err := newCommand("nerd version", "check the current version", "", nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create command")
		}
		cmd := &Version{
			version: version,
			commit:  commit,
			command: comm,
		}

		return cmd, nil
	}
}

//Run is called by run and allows an error to be returned
func (cmd *Version) Run(args []string) int {
	cmd.ui.Info(fmt.Sprintf("%s (%s)", cmd.version, cmd.commit))
	nerd.VersionMessage(cmd.version)
	return 0
}
