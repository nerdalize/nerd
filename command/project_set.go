package command

import (
	"net/url"
	"os"
	"path/filepath"

	"github.com/mitchellh/cli"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/nerdalize/nerd/nerd/client/auth/v1"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/nerd/oauth"
	"github.com/nerdalize/nerd/pkg/populator"
	"github.com/pkg/errors"
)

//ProjectSetOpts determine
type ProjectSetOpts struct {
	Config string `long:"config-src" default:"env" default-mask:"" description:"type of configuration to use (from env, endpoint, or oidc)"`
}

//ProjectSet command
type ProjectSet struct {
	*command
	opts *ProjectSetOpts
}

//ProjectSetFactory returns a factory method for the join command
func ProjectSetFactory() (cli.Command, error) {
	opts := &ProjectSetOpts{}
	comm, err := newCommand("nerd project set", "Set current working project.", "", opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &ProjectSet{
		command: comm,
		opts:    opts,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *ProjectSet) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return errShowHelp("Not enough arguments, see below for usage.")
	}
	projectSlug := args[0]

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

	// This part should be easier to do if we can get a project with its name instead of its id
	projects, err := client.ListProjects()
	if err != nil {
		return HandleError(err)
	}

	ok := false
	for _, project := range projects.Projects {
		if project.Slug == args[0] {
			ok = true
		}
	}

	if !ok {
		return HandleError(errors.New("Project not found, please check the project name. You can get a list of your projects by running `nerd project list`."))
	}

	var kubeConfigFile string
	if os.Getenv("KUBECONFIG") == "" {
		hdir, err := homedir.Dir()
		if err != nil {
			return HandleError(err)
		}
		kubeConfigFile = filepath.Join(hdir, ".kube", "config")
	} else {
		kubeConfigFile = filepath.Join(os.Getenv("KUBECONFIG"), "config")
	}
	p, err := populator.New(cmd.opts.Config, kubeConfigFile)
	if err != nil {
		return HandleError(err)
	}
	err = p.PopulateKubeConfig(args[0])
	if err != nil {
		return HandleError(err)
	}

	err = cmd.session.WriteProject(projectSlug, conf.DefaultAWSRegion)
	if err != nil {
		return HandleError(err)
	}

	return nil
}
