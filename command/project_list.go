package command

import (
	"fmt"
	"net/url"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/nerd/oauth"
	"github.com/pkg/errors"
)

//ProjectListOps describes command options
type ProjectListOps struct {
	NerdOpts
}

//ProjectList command
type ProjectList struct {
	*command
	opts   *ProjectListOps
	parser *flags.Parser
}

//ProjectListFactory returns a factory method for the join command
func ProjectListFactory() (cli.Command, error) {
	cmd := &ProjectList{
		command: &command{
			help:     "",
			synopsis: "list all your projects",
			parser:   flags.NewNamedParser("nerd project list", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &ProjectListOps{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *ProjectList) DoRun(args []string) (err error) {
	config, err := conf.Read()
	if err != nil {
		HandleError(err)
	}
	authbase, err := url.Parse(config.Auth.APIEndpoint)
	if err != nil {
		HandleError(errors.Wrapf(err, "auth endpoint '%v' is not a valid URL", config.Auth.APIEndpoint))
	}
	authOpsClient := v1auth.NewOpsClient(v1auth.OpsClientConfig{
		Base:   authbase,
		Logger: logrus.StandardLogger(),
	})
	client := v1auth.NewClient(v1auth.ClientConfig{
		Base:               authbase,
		Logger:             logrus.StandardLogger(),
		OAuthTokenProvider: oauth.NewConfigProvider(authOpsClient),
	})

	projects, err := client.ListProjects()
	if err != nil {
		HandleError(err)
	}
	for _, project := range projects.Projects {
		fmt.Println("-", project.Code)
	}

	return nil
}
