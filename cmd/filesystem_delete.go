package cmd

import (
	"context"
	"fmt"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
)

//FileSystemDelete command
type FileSystemDelete struct {
	All bool `long:"all" short:"a" description:"delete all your file systems in one command"`

	*command
}

//FileSystemDeleteFactory creates the command
func FileSystemDeleteFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &FileSystemDelete{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, nil, flags.None, "nerd fs delete")
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *FileSystemDelete) Execute(args []string) (err error) {
	if cmd.All {
		return cmd.deleteAll()
	}
	if len(args) < 1 {
		return errShowUsage(fmt.Sprintf(MessageNotEnoughArguments, 1, ""))
	}

	kopts := cmd.globalOpts.KubeOpts
	deps, err := NewDeps(cmd.Logger(), kopts)
	if err != nil {
		return renderConfigError(err, "failed to configure")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, kopts.Timeout)
	defer cancel()

	kube := svc.NewKube(deps)
	for i := range args {
		in := &svc.DeleteFileSystemInput{
			Name: args[i],
		}

		_, err = kube.DeleteFileSystem(ctx, in)
		if err != nil {
			return renderServiceError(err, fmt.Sprintf("failed to delete file system `%s`", in.Name))
		}

		cmd.out.Infof("Deleted file system: '%s'", in.Name)
	}
	return nil
}

func (cmd *FileSystemDelete) deleteAll() error {
	kopts := cmd.globalOpts.KubeOpts
	deps, err := NewDeps(cmd.Logger(), kopts)
	if err != nil {
		return renderConfigError(err, "failed to configure")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, kopts.Timeout)
	defer cancel()

	s, err := cmd.out.Ask("Are you sure you want to delete all your file systems? (y/N)")
	if err != nil {
		return err
	}
	if !strings.HasPrefix(strings.ToLower(s), "y") {
		return nil
	}

	kube := svc.NewKube(deps)
	filesystems, err := kube.ListFileSystems(ctx, &svc.ListFileSystemsInput{})
	if err != nil {
		return renderServiceError(err, "failed to get all file systems")
	}
	if len(filesystems.Items) == 0 {
		cmd.out.Info("No file system found.")
	}
	for _, fs := range filesystems.Items {
		in := &svc.DeleteFileSystemInput{
			Name: fs.Name,
		}

		_, err = kube.DeleteFileSystem(ctx, in)
		if err != nil {
			return renderServiceError(err, fmt.Sprintf("failed to delete file system `%s`", in.Name))
		}

		cmd.out.Infof("Deleted file system: '%s'", in.Name)
	}
	return nil
}

// Description returns long-form help text
func (cmd *FileSystemDelete) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *FileSystemDelete) Synopsis() string { return "Remove a file system from the cluster." }

// Usage shows usage
func (cmd *FileSystemDelete) Usage() string { return "nerd fs delete FILE_SYSTEM_NAME [FILES_YSTEM_NAME...]" }
