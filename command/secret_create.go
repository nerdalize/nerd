package command

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//SecretCreate command
type SecretCreate struct {
	*command
}

//SecretCreateFactory returns a factory method for the join command
func SecretCreateFactory() (cli.Command, error) {
	comm, err := newCommand("nerd secret create <name> <key> <value>", "create secrets to be used by workers", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &SecretCreate{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *SecretCreate) DoRun(args []string) (err error) {
	if len(args) < 3 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	bclient, err := NewClient(cmd.ui, cmd.config, cmd.session)
	if err != nil {
		HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}

	out, err := bclient.CreateSecret(ss.Project.Name, args[0], args[1], args[2])
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Secret Creation: %v", out)
	return nil
}
