package command

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//SecretDescribe command
type SecretDescribe struct {
	*command
}

//SecretDescribeFactory returns a factory method for the join command
func SecretDescribeFactory() (cli.Command, error) {
	comm, err := newCommand("nerd secret describe <name>", "show more information about a specific secret", "", nil)
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
		return fmt.Errorf("not enough arguments, see --help")
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}
	out, err := bclient.DescribeSecret(ss.Project.Name, args[0])
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Secret Description: %+v", out)
	return nil
}
