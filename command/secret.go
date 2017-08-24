package command

import (
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

// Secret command
type Secret struct {
	*command
}

var synopsisSecret = "Set and list secrets. A secret can be credentials for a Docker registry, or other sensitive information needed by your project."

// SecretFactory returns a factory method for the secret command
func SecretFactory() (cli.Command, error) {
	comm, err := newCommand("nerd secret <subcommand> {opaque, registry}", synopsisSecret, "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &Project{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *Secret) DoRun(args []string) (err error) {
	return errShowHelp("Not enough arguments, see below for usage.")
}
