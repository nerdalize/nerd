package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/cli"
	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
)

// SecretCreateOpts describes the options to the SecretCreate command
type SecretCreateOpts struct {
	Username string `long:"username" default:"" default-mask:"" description:"Username for Docker registry authentication"`
	Password string `long:"password" default:"" default-mask:"" description:"Password for Docker registry authentication"`
	Type     string `long:"type" default:"opaque" default-mask:"" description:"Type of secret to display, defaults to opaque."`
}

//SecretCreate command
type SecretCreate struct {
	*command
	opts *SecretCreateOpts
}

//SecretCreateFactory returns a factory method for the secret create command
func SecretCreateFactory() (cli.Command, error) {
	opts := &SecretCreateOpts{}
	comm, err := newCommand("nerd secret create <name> [key=val]", "Create secrets to be used by workers.", "", opts)
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
	if len(args) < 1 {
		return errors.Wrap(errShowHelp("show help"), "Not enough arguments, see below for usage.")
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}

	secretName := args[0]
	var out *v1payload.CreateSecretOutput
	if cmd.opts.Type == v1payload.SecretTypeRegistry {
		out, err = bclient.CreatePullSecret(ss.Project.Name,
			secretName,
			cmd.opts.Username,
			cmd.opts.Password,
		)
		if err != nil {
			return HandleError(err)
		}
	} else if cmd.opts.Type == v1payload.SecretTypeOpaque {
		if len(args) < 2 {
			return HandleError(fmt.Errorf("provide a valid key value pair: key=value"))
		}
		secretKv := strings.Split(args[1], "=")
		if len(secretKv) < 2 {
			return HandleError(fmt.Errorf("provide a valid key value pair (key=value)"))
		}
		out, err = bclient.CreateSecret(ss.Project.Name, secretName, secretKv[0], secretKv[1])
		if err != nil {
			return HandleError(err)
		}
	} else {
		return HandleError(fmt.Errorf("invalid secret type '%s', available options are 'registry', and 'opaque'", cmd.opts.Type))
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Type"})
	row := []string{}
	row = append(row, out.Name)
	row = append(row, out.Type)
	table.Append(row)

	table.Render()
	return nil
}
