package cmd

import (
	"fmt"
	"net/url"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	homedir "github.com/mitchellh/go-homedir"
	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	v1authpayload "github.com/nerdalize/nerd/nerd/client/auth/v1/payload"
	"github.com/nerdalize/nerd/nerd/oauth"
	"github.com/nerdalize/nerd/pkg/populator"
	"github.com/pkg/errors"
)

//ClusterSet command
type ClusterSet struct {
	Namespace string `long:"namespace" short:"n" description:"set a specific namespace as the default one"`

	*command
}

//ClusterSetFactory creates the command
func ClusterSetFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &ClusterSet{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, &ConfOpts{}, flags.None, "nerd cluster set")
	t, ok := cmd.advancedOpts.(*ConfOpts)
	if !ok {
		return nil
	}
	t.ConfigFile = cmd.setConfig
	t.SessionFile = cmd.setSession
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *ClusterSet) Execute(args []string) (err error) {
	if len(args) > 1 {
		return errShowUsage(fmt.Sprintf(MessageTooManyArguments, 1, ""))
	} else if len(args) < 1 {
		return errShowUsage(fmt.Sprintf(MessageNotEnoughArguments, 1, ""))
	}
	name := args[0]

	authbase, err := url.Parse(cmd.config.Auth.APIEndpoint)
	if err != nil {
		return errors.Wrapf(err, "auth endpoint '%v' is not a valid URL", cmd.config.Auth.APIEndpoint)
	}
	authOpsClient := v1auth.NewOpsClient(v1auth.OpsClientConfig{
		Base:   authbase,
		Logger: cmd.Logger(),
	})
	client := v1auth.NewClient(v1auth.ClientConfig{
		Base:               authbase,
		Logger:             cmd.Logger(),
		OAuthTokenProvider: oauth.NewConfigProvider(authOpsClient, cmd.config.Auth.SecureClientID, cmd.config.Auth.SecureClientSecret, cmd.session),
	})

	list, err := client.ListClusters()
	if err != nil {
		return err
	}

	var cluster *v1authpayload.GetClusterOutput
	for _, c := range list.Clusters {
		if c.ShortName == name {
			cluster, err = client.GetCluster(c.URL)
			if err != nil {
				return err
			}
		}
	}

	c := populator.Client{
		Secret:       cmd.config.Auth.SecureClientSecret,
		ID:           cmd.config.Auth.SecureClientID,
		IDPIssuerURL: cmd.config.Auth.IDPIssuerURL,
	}

	hdir, err := homedir.Dir()
	if err != nil {
		return err
	}
	kubeConfig := cmd.globalOpts.KubeOpts.KubeConfig
	if kubeConfig == "" {
		kubeConfig = filepath.Join(hdir, ".kube", "config")
	}
	p, err := populator.New(&c, "generic", kubeConfig, hdir, cluster)
	if err != nil {
		return err
	}
	err = p.PopulateKubeConfig(cmd.Namespace)
	if err != nil {
		p.RemoveConfig(cluster.ShortName)
		return err
	}
	if err := checkNamespace(kubeConfig, cmd.Namespace); err != nil {
		p.RemoveConfig(cluster.ShortName)
		return err
	}

	cmd.out.Infof("You are now using '%s' config.", name)
	return nil
}

// Description returns long-form help text
func (cmd *ClusterSet) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *ClusterSet) Synopsis() string { return "Set a specific cluster as the current one to use." }

// Usage shows usage
func (cmd *ClusterSet) Usage() string { return "nerd cluster set NAME [OPTIONS]" }
