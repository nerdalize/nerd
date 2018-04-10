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

	var clusters []string
	for key := range config.Clusters {
		clusters = append(clusters, key)
	}
	sort.Strings(clusters)

	hdr := []string{"CLUSTER"}
	rows := [][]string{}
	for _, name := range clusters {
		rows = append(rows, []string{
			name,
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
