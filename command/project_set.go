package command

import (
	"net/url"
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
	Config     string `long:"config-src" default:"env" default-mask:"" description:"type of configuration to use (from env, endpoint, or oidc)"`
	KubeConfig string `long:"kube-config" env:"KUBECONFIG" description:"file at which Nerd will look for Kubernetes credentials" default-mask:"~/.kube/conf"`
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

	project, err := client.GetProject(projectSlug)
	if err != nil {
		return HandleError(errors.Wrap(err, "Project not found, please check the project name. You can get a list of your projects by running `nerd project list`."))
	}

	if cmd.opts.KubeConfig == "" {
		hdir, err := homedir.Dir()
		if err != nil {
			return HandleError(err)
		}
		cmd.opts.KubeConfig = filepath.Join(hdir, ".kube", "config")
	}
	p, err := populator.New(cmd.opts.Config, cmd.opts.KubeConfig, project)
	if err != nil {
		return HandleError(err)
	}
	err = p.PopulateKubeConfig(projectSlug)
	if err != nil {
		return HandleError(err)
	}

	err = cmd.session.WriteProject(projectSlug, conf.DefaultAWSRegion)
	if err != nil {
		return HandleError(err)
	}

	cmd.outputter.Logger.Printf("Project %s set successfully", projectSlug)
	return nil
}
