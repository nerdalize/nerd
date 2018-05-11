package cmd

import (
	"fmt"
	"net/url"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	"github.com/nerdalize/nerd/nerd/oauth"
	"github.com/pkg/errors"
)

//ClusterList command
type ClusterList struct {
	*command
}

//ClusterListFactory creates the command
func ClusterListFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &ClusterList{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, nil, flags.None, "nerd cluster list")
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *ClusterList) Execute(args []string) (err error) {
	if len(args) > 0 {
		return errShowUsage(MessageNoArgumentRequired)
	}

	// TODO
	authbase, err := url.Parse(cmd.config.Auth.APIEndpoint)
	if err != nil {
		return errors.Wrapf(err, "auth endpoint '%v' is not a valid URL", cmd.config.Auth.APIEndpoint)
	}
	authOpsClient := v1auth.NewOpsClient(v1auth.OpsClientConfig{
		Base: authbase,
	})
	client := v1auth.NewClient(v1auth.ClientConfig{
		Base:               authbase,
		OAuthTokenProvider: oauth.NewConfigProvider(authOpsClient, cmd.config.Auth.SecureClientID, cmd.config.Auth.SecureClientSecret, cmd.session),
	})

	clusters, err := client.ListClusters()
	if err != nil {
		return err
	}

	// Add role (admin, team member ...)
	hdr := []string{"CLUSTER", "CPU", "MEMORY", "PODS"}
	rows := [][]string{}
	for _, cluster := range clusters.Clusters {
		rows = append(rows, []string{
			cluster.Name,
			fmt.Sprintf("%s/%s", cluster.Usage.CPU, cluster.Capacity.CPU),
			fmt.Sprintf("%s/%s", cluster.Usage.Memory, cluster.Capacity.Memory),
			fmt.Sprintf("%d/%d", cluster.Usage.Pods, cluster.Capacity.Pods),
		})
	}

	return cmd.out.Table(hdr, rows)
}

// Description returns long-form help text
func (cmd *ClusterList) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *ClusterList) Synopsis() string { return "Returns an overview of your clusters." }

// Usage shows usage
func (cmd *ClusterList) Usage() string { return "nerd cluster list [OPTIONS]" }
