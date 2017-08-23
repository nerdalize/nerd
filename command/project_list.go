package command

import (
	"net/url"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command/format"
	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	"github.com/nerdalize/nerd/nerd/oauth"
	"github.com/pkg/errors"
)

//ProjectList command
type ProjectList struct {
	*command
}

//ProjectListFactory returns a factory method for the join command
func ProjectListFactory() (cli.Command, error) {
	comm, err := newCommand("nerd project list", "List all your projects.", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &ProjectList{
		command: comm,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *ProjectList) DoRun(args []string) (err error) {
	authbase, err := url.Parse(cmd.config.Auth.APIEndpoint)
	if err != nil {
		return errors.Wrapf(err, "auth endpoint '%v' is not a valid URL", cmd.config.Auth.APIEndpoint)
	}
	authOpsClient := v1auth.NewOpsClient(v1auth.OpsClientConfig{
		Base:   authbase,
		Logger: cmd.outputter.Logger,
	})
	client := v1auth.NewClient(v1auth.ClientConfig{
		Base:               authbase,
		Logger:             cmd.outputter.Logger,
		OAuthTokenProvider: oauth.NewConfigProvider(authOpsClient, cmd.config.Auth.ClientID, cmd.session),
	})

	projects, err := client.ListProjects()
	if err != nil {
		return errors.Wrap(err, "failed to list projects")
	}

	header := "Nr\tProject ID"
	pretty := "{{range $i, $x := $.Projects}}{{$x.ID}}\t{{$x.Slug}}\n{{end}}"
	raw := "{{range $i, $x := $.Projects}}{{$x.ID}}\t{{$x.Slug}}\t{{$x.URL}}\n{{end}}"
	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(projects, header, pretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(projects, raw),
		format.OutputTypeJSON:   format.NewJSONDecorator(projects.Projects),
	})

	return nil
}
