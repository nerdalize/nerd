package cmd

import (
	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//FileSystem command
type FileSystem struct {
	*command
}

//FileSystemFactory creates the command
func FileSystemFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &FileSystem{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, nil, flags.None, "nerd fs")

	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *FileSystem) Execute(args []string) (err error) { return errShowHelp("") }

// Description returns long-form help text
func (cmd *FileSystem) Description() string {
	return "Group of commands used to manage file systems. A file system is a network folder that can be freely accessed by pods. It can be used to provide shared input data, a space to write results, or just for cloud storage in general."
}

// Synopsis returns a one-line
func (cmd *FileSystem) Synopsis() string {
	return "Group of commands used to manage file systems (network folders)."
}

// Usage shows usage
func (cmd *FileSystem) Usage() string { return "nerd fs <subcommand>" }
