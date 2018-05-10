package cmd

import (
	"fmt"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	homedir "github.com/mitchellh/go-homedir"
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
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, nil, flags.None, "nerd cluster set")
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

	// TODO
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

	err = populator.UseConfig(name, path)
	if err != nil {
		return renderServiceError(err, "failed to use precised config")
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
