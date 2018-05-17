package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	homedir "github.com/mitchellh/go-homedir"
	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	v1authpayload "github.com/nerdalize/nerd/nerd/client/auth/v1/payload"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/nerd/oauth"
	"github.com/nerdalize/nerd/pkg/populator"
	"github.com/pkg/errors"
)

//ClusterSetDefault command
type ClusterSetDefault struct {
	Namespace string `long:"namespace" short:"n" description:"set a specific namespace as the default one"`

	*command
}

//ClusterSetDefaultFactory creates the command
func ClusterSetDefaultFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &ClusterSetDefault{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, &ConfOpts{}, flags.None, "nerd cluster set-default")
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
func (cmd *ClusterSetDefault) Execute(args []string) (err error) {
	// TODO move this part to another func
	env := os.Getenv("NERD_ENV")
	if env == "staging" {
		cmd.config = conf.StagingDefaults()
	} else if env == "dev" {
		cmd.config = conf.DevDefaults(os.Getenv("NERD_API_ENDPOINT"))
	}

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
		if c.Name == name {
			cluster, err = client.GetCluster(c.URL)
			if err != nil {
				return err
			}
		}
	}

	if cluster == nil {
		cmd.out.Infof("Cluster not found. You can use `nerd cluster list` to see your clusters.")
		return nil
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
	if cmd.Namespace == "" && len(cluster.Namespaces) >= 1 {
		cmd.Namespace = cluster.Namespaces[0].Name
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
func (cmd *ClusterSetDefault) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *ClusterSetDefault) Synopsis() string {
	return "Set a specific cluster as the current one to use."
}

// Usage shows usage
func (cmd *ClusterSetDefault) Usage() string { return "nerd cluster set-default NAME [OPTIONS]" }
