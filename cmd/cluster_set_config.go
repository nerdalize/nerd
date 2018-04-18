package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd/api"

	"github.com/nerdalize/nerd/pkg/populator"

	"github.com/mitchellh/cli"
	homedir "github.com/mitchellh/go-homedir"
)

//ClusterSetConfig command options
type ClusterSetConfig struct {
	populator.Context
	populator.Cluster
	populator.Auth
	*command
}

//ClusterSetConfigFactory creates the command
func ClusterSetConfigFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &ClusterSetConfig{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, nil, flags.PassAfterNonOption, "nerd cluster set-config")
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *ClusterSetConfig) Execute(args []string) (err error) {
	if len(args) < 1 {
		return errShowUsage(fmt.Sprintf(MessageNotEnoughArguments, 1, ""))
	} else if len(args) > 1 {
		return errShowUsage(fmt.Sprintf(MessageTooManyArguments, 1, ""))
	}

	name := args[0]
	configPath := cmd.globalOpts.KubeOpts.KubeConfig

	//Expand tilde for homedir
	configPath, err = homedir.Expand(configPath)
	if err != nil {
		return errors.Wrap(err, "failed to expand home directory in kubeconfig local path")
	}

	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return renderServiceError(err, "failed to turn local path into absolute path")
	}

	config, err := populator.ReadConfigOrNew(configPath)
	if err != nil {
		return renderConfigError(err, "failed to read kubeconfig")
	}

	if cmd.Context.Namespace != "" && cmd.Context.Server != "" && cmd.Context.User != "" {
		config.Contexts[name] = &api.Context{
			Namespace: cmd.Context.Namespace,
			Cluster:   cmd.Context.Server,
			AuthInfo:  cmd.Context.User,
		}
	}

	config.CurrentContext = name
	err = populator.WriteConfig(config, configPath)
	if err != nil {
		return renderServiceError(err, "failed to set config")
	}

	cmd.out.Infof("Configured cluster: '%s'", args[0])
	cmd.out.Infof("To run a job with a cluster, use: 'nerd job run'")
	return nil
}

// Description returns long-form help text
func (cmd *ClusterSetConfig) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *ClusterSetConfig) Synopsis() string { return "Set a config to use your compute cluster." }

// Usage shows usage
func (cmd *ClusterSetConfig) Usage() string {
	return "nerd cluster set-config [OPTIONS] CONFIG-NAME"
}
