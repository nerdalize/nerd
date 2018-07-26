package cmd

import (
    "context"
    "fmt"

    "github.com/jessevdk/go-flags"
    "github.com/mitchellh/cli"

    "github.com/nerdalize/nerd/svc"
)

//FileSystemCreate command
type FileSystemCreate struct {
    Name       string   `long:"name" short:"n" description:"assign a name to the file system"`
    *command
}

//FileSystemCreateFactory creates the command
func FileSystemCreateFactory(ui cli.Ui) cli.CommandFactory {
    cmd := &FileSystemCreate{}
    cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, nil, flags.PassAfterNonOption, "nerd fs create")
    return func() (cli.Command, error) {
        return cmd, nil
    }
}

//Execute runs the command
func (cmd *FileSystemCreate) Execute(args []string) (err error) {
    if len(args) < 1 {
        return errShowUsage(fmt.Sprintf(MessageNotEnoughArguments, 1, ""))
    }

    kopts := cmd.globalOpts.KubeOpts
    deps, err := NewDeps(cmd.Logger(), kopts)
    if err != nil {
        return renderConfigError(err, "failed to configure")
    }

    kube := svc.NewKube(deps)

    ctx := context.Background()
    ctx, cancel := context.WithTimeout(ctx, kopts.Timeout)
    defer cancel()

    out, err := kube.CreateFileSystem(ctx, &svc.CreateFileSystemInput{
        Name: cmd.Name,
        Capacity: args[0],
    })

    if err != nil {
        return renderServiceError(err, "failed to create persistent volume claim")
    }

    cmd.out.Infof("Started creation of file system: '%s'", out.Name)
    cmd.out.Infof("To see when it's ready, use: 'nerd fs list'")

    return nil
}

// Description returns long-form help text
func (cmd *FileSystemCreate) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *FileSystemCreate) Synopsis() string { return "Create a file system on your compute cluster." }

// Usage shows usage
func (cmd *FileSystemCreate) Usage() string { return "nerd fs create [OPTIONS] SIZE" }
