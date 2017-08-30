package command

import (
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

// SecretDelete command
type SecretDelete struct {
	*command
}

// SecretDeleteFactory returns a factory method for the secret delete command
func SecretDeleteFactory() (cli.Command, error) {
	comm, err := newCommand("nerd secret delete <name>", "Remove a secret.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &SecretDelete{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *SecretDelete) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return errShowHelp("Not enough arguments, see below for usage.")
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}
	_, err = ss.RequireProjectID()
	if err != nil {
		return HandleError(err)
	}

	_, err = bclient.DeleteSecret(ss.Project.Name, args[0])
	if err != nil {
		return HandleError(err)
	}

	cmd.outputter.Logger.Printf("Secret successfully deleted")
	return nil
}
