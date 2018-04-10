package cmd

import (
	"path/filepath"
	"sort"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/nerdalize/nerd/pkg/populator"
	"github.com/pkg/errors"
)

//ClusterListConfig command
type ClusterListConfig struct {
	*command
}

//ClusterListConfigFactory creates the command
func ClusterListConfigFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &ClusterListConfig{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, nil, flags.None, "nerd cluster list-config")
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *ClusterListConfig) Execute(args []string) (err error) {
	if len(args) > 0 {
		return errShowUsage(MessageNoArgumentRequired)
	}

	var path string
	//Expand tilde for homedir
	path, err = homedir.Expand(cmd.globalOpts.KubeConfig)
	if err != nil {
		return errors.Wrap(err, "failed to expand home directory in kubeconfig local path")
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return renderServiceError(err, "failed to turn local path into absolute path")
	}

	config, err := populator.ReadConfig(path)
	if err != nil {
		return renderConfigError(err, "failed to read kubeconfig")
	}
	// To store the keys in slice in sorted order
	var keys []string
	for k := range config.Contexts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	hdr := []string{"CONFIG", "CLUSTER", "USER", "NAMESPACE"}
	rows := [][]string{}
	for _, key := range keys {
		rows = append(rows, []string{
			key,
			config.Contexts[key].Cluster,
			config.Contexts[key].AuthInfo,
			config.Contexts[key].Namespace,
		})
	}
	return cmd.out.Table(hdr, rows)
}

// Description returns long-form help text
func (cmd *ClusterListConfig) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *ClusterListConfig) Synopsis() string { return "Returns an overview of your cluster configs." }

// Usage shows usage
func (cmd *ClusterListConfig) Usage() string { return "nerd cluster list-config [OPTIONS]" }
