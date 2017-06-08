package command

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
	"github.com/pkg/errors"
)

// SecretCreateOpts describes the options to the SecretCreate command
type SecretCreateOpts struct {
	DockerServer   string `long:"docker-server" default:"" default-mask:"" description:"https://index.docker.io/v1/': Server location for Docker registry"`
	DockerUsername string `long:"docker-username" default:"" default-mask:"" description:"Username for Docker registry authentication"`
	DockerPassword string `long:"docker-password" default:"" default-mask:"" description:"Password for Docker registry authentication"`
	DockerEmail    string `long:"docker-email" default:"" default-mask:"" description:"email for Docker registry"`
}

//SecretCreate command
type SecretCreate struct {
	*command
	opts *SecretCreateOpts
}

//SecretCreateFactory returns a factory method for the join command
func SecretCreateFactory() (cli.Command, error) {
	opts := &SecretCreateOpts{}
	comm, err := newCommand("nerd secret create <name> [key] [value]", "create secrets to be used by workers", "", opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &SecretCreate{
		command: comm,
		opts:    opts,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *SecretCreate) DoRun(args []string) (err error) {
	isPullSecret := cmd.opts.DockerUsername != "" && cmd.opts.DockerPassword != ""

	if len(args) < 3 && !isPullSecret {
		return fmt.Errorf("not enough arguments, see --help")
	} else if len(args) < 1 {
		return fmt.Errorf("please provide a pull secret name")
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}

	var out *v1payload.CreateSecretOutput
	if isPullSecret {
		out, err = bclient.CreatePullSecret(ss.Project.Name,
			args[0],
			cmd.opts.DockerServer,
			cmd.opts.DockerUsername,
			cmd.opts.DockerPassword,
			cmd.opts.DockerEmail)
		if err != nil {
			HandleError(err)
		}
	} else {
		out, err = bclient.CreateSecret(ss.Project.Name, args[0], args[1], args[2])
		if err != nil {
			HandleError(err)
		}
	}

	logrus.Infof("Secret Creation: %v", out)
	return nil
}
