package command

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/client"
	"github.com/nerdalize/nerd/nerd/client/credentials/provider"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

//LoginOpts describes command options
type LoginOpts struct {
	User string `long:"user" default:"" default-mask:"" env:"NERD_USER" description:"nerd username"`
	Pass string `long:"pass" default:"" default-mask:"" env:"NERD_PASS" description:"nerd password"`
	NerdOpts
}

//Login command
type Login struct {
	*command

	opts   *LoginOpts
	parser *flags.Parser
}

//LoginFactory returns a factory method for the join command
func LoginFactory() func() (cmd cli.Command, err error) {
	cmd := &Login{
		command: &command{
			help:     "",
			synopsis: "Setup an authorized session.",
			parser:   flags.NewNamedParser("nerd login", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &LoginOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//DoRun is called by run and allows an error to be returned
func (cmd *Login) DoRun(args []string) error {
	config, errRead := conf.Read()
	if errRead != nil {
		return errors.Wrap(errRead, "failed to read nerd config file")
	}

	oauthAPI := provider.NewOAuthAPI(client.NewAuthAPI(config.Auth))

	_, err := oauthAPI.RetrieveWithoutKey()
	if err != nil {
		return errors.Wrap(err, "failed to fetch oauth tokens")
	}

	logrus.Info("Successful login")
	return nil
}
