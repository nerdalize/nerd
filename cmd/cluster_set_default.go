package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	v1authpayload "github.com/nerdalize/nerd/nerd/client/auth/v1/payload"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/nerd/oauth"
	"github.com/nerdalize/nerd/pkg/kubeconfig"
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
	cluster, err = lookByID(args[0], list.Clusters)
	if err != nil {
		cluster, err = lookByName(args[0], list.Clusters)
		if err != nil {
			return err
		}
	}
	cluster, err = client.GetCluster(cluster.URL)
	if err != nil {
		return err
	}
	if cluster == nil {
		cmd.out.Infof("Cluster not found. You can use `nerd cluster list` to see your clusters.")
		return nil
	}

	kubeConfig, err := kubeconfig.GetPath(cmd.globalOpts.KubeConfig)
	if err != nil {
		return err
	}
	p, err := populator.New("generic", kubeConfig, cluster)
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

	// check if it's nerd ready (is app nlz-utils up ?)
	// else apply helm chart
	// - get catalogs (is it necessary ?)
	// - get projects (can we have it from auth)
	// - launch app
	// how to get the right chart version ?
	name := cluster.Name
	if name == "" {
		name = cluster.ShortName
	}
	cmd.out.Infof("You are now using %s's config.", name)
	return nil
}

func lookByID(name string, clusters []*v1authpayload.GetClusterOutput) (*v1authpayload.GetClusterOutput, error) {
	id, err := strconv.Atoi(name)
	if err != nil {
		return nil, err
	}
	id--
	if len(clusters) >= id {
		return clusters[id], nil
	}
	return nil, errors.New("cluster not found")
}

func lookByName(name string, clusters []*v1authpayload.GetClusterOutput) (*v1authpayload.GetClusterOutput, error) {
	for _, cluster := range clusters {
		if cluster.ShortName == name || cluster.Name == name {
			return cluster, nil
		}
	}
	return nil, errors.New("cluster not found")
}

// Description returns long-form help text
func (cmd *ClusterSetDefault) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *ClusterSetDefault) Synopsis() string {
	return "Set a specific cluster as the current one to use."
}

// Usage shows usage
func (cmd *ClusterSetDefault) Usage() string { return "nerd cluster set-default NAME [OPTIONS]" }
