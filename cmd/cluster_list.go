package cmd

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	v1payload "github.com/nerdalize/nerd/nerd/client/auth/v1/payload"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/nerd/oauth"
	"github.com/nerdalize/nerd/pkg/kubeconfig"
	"github.com/nerdalize/nerd/pkg/populator"
	"github.com/pkg/errors"
)

//ClusterList command
type ClusterList struct {
	*command
}

//ClusterListFactory creates the command
func ClusterListFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &ClusterList{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, &ConfOpts{}, flags.None, "nerd cluster list")
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
func (cmd *ClusterList) Execute(args []string) (err error) {
	// TODO move this part to another func
	env := os.Getenv("NERD_ENV")
	if env == "staging" {
		cmd.config = conf.StagingDefaults()
	} else if env == "dev" {
		cmd.config = conf.DevDefaults(os.Getenv("NERD_API_ENDPOINT"))
	}

	if len(args) > 0 {
		return errShowUsage(MessageNoArgumentRequired)
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
		OAuthTokenProvider: oauth.NewConfigProvider(authOpsClient, cmd.config.Auth.SecureClientID, cmd.config.Auth.SecureClientSecret, cmd.session),
		Logger:             cmd.Logger(),
	})

	kubeConfig, err := kubeconfig.GetPath(cmd.globalOpts.KubeConfig)
	if err != nil {
		return err
	}
	conf, err := populator.ReadConfigOrNew(kubeConfig)
	if err != nil {
		return err
	}

	clusters, err := client.ListClusters()
	if err != nil {
		return renderClientError(err, "failed to get your cluster list")
	}

	// TODO Add role (admin, team member ...)
	hdr := []string{"ID", "CLUSTER NAME", "VCPUS", "MEMORY", "PODS"}
	rows := [][]string{}
	for x, cluster := range clusters.Clusters {
		if cluster.Name == "" {
			cluster.Name = cluster.ShortName
		}
		if clusterMatchContext(cluster, conf.CurrentContext) {
			cluster.Name = fmt.Sprintf("%s *", cluster.Name)
		}
		id := strconv.Itoa(x + 1)
		rows = append(rows, []string{
			id,
			cluster.Name,
			fmt.Sprintf("%.1f/%.1f", cluster.Usage.CPU, cluster.Capacity.CPU),
			fmt.Sprintf("%s/%s", renderFloatToMem(cluster.Usage.Memory), renderFloatToMem(cluster.Capacity.Memory)),
			fmt.Sprintf("%d/%d", cluster.Usage.Pods, cluster.Capacity.Pods),
		})
	}

	return cmd.out.Table(hdr, rows)
}

func renderFloatToMem(n float64) string {
	return fmt.Sprintf("%.1f", n/1000/1000/1000)
}

func clusterMatchContext(cluster *v1payload.GetClusterOutput, context string) bool {
	if len(cluster.Namespaces) > 0 {
		for x := range cluster.Namespaces {
			if fmt.Sprintf("%s-%s-%s", populator.Prefix, cluster.ShortName, cluster.Namespaces[x].Name) == context {
				return true
			}
		}
	}
	return fmt.Sprintf("%s-%s", populator.Prefix, cluster.ShortName) == context
}

// Description returns long-form help text
func (cmd *ClusterList) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *ClusterList) Synopsis() string { return "Returns an overview of your clusters." }

// Usage shows usage
func (cmd *ClusterList) Usage() string { return "nerd cluster list [OPTIONS]" }
