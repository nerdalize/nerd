package cmd

import (
    "context"
    "fmt"
    "strings"

    flags "github.com/jessevdk/go-flags"
    "github.com/mitchellh/cli"
    "github.com/skratchdot/open-golang/open"
    
    "github.com/nerdalize/nerd/svc"
)

//FileSystemMount command
type FileSystemMount struct {
    *command
}

//FileSystemMountFactory creates the command
func FileSystemMountFactory(ui cli.Ui) cli.CommandFactory {
    cmd := &FileSystemMount{}
    cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, nil, flags.None, "nerd fs mount")
    return func() (cli.Command, error) {
        return cmd, nil
    }
}

//Execute runs the command
func (cmd *FileSystemMount) Execute(args []string) (err error) {
    if len(args) < 1 {
        return errShowUsage(fmt.Sprintf(MessageNotEnoughArguments, 1, ""))
    } else if len(args) > 1 {
        return errShowUsage(fmt.Sprintf(MessageTooManyArguments, 1, ""))
    }

    kopts := cmd.globalOpts.KubeOpts
    deps, err := NewDeps(cmd.Logger(), kopts)
    if err != nil {
        return renderConfigError(err, "failed to configure")
    }

    ctx := context.Background()
    ctx, cancel := context.WithTimeout(ctx, kopts.Timeout)
    defer cancel()

    in := &svc.GetFileSystemInput{
        Name: args[0],
    }
    kube := svc.NewKube(deps)
    out, err := kube.GetFileSystem(ctx, in)
    if err != nil {
        return renderServiceError(err, "failed to get file system")
    }

    inPVC := &svc.GetPersistentVolumeInput{
        Name: out.VolumeName,
    }
    outPVC, err := kube.GetPersistentVolume(ctx, inPVC)
    if err != nil {
        return renderServiceError(err, "failed to get volume")
    }

    // Display mount info
    cmd.out.Infof("The file system '%s' can be accessed through WebDAV at http://%s:%d%s.", out.Name, outPVC.WebDAVHost, outPVC.WebDAVPort, outPVC.WebDAVPath)

    // Try to open share automatically (for Windows users)
    windowsPath := strings.Replace(outPVC.WebDAVPath, "/", "\\", -1)
    open.RunWith(fmt.Sprintf("\\\\%s@%d%s", outPVC.WebDAVHost, outPVC.WebDAVPort, windowsPath), "explorer")

    return nil
}

// Description returns long-form help text
func (cmd *FileSystemMount) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *FileSystemMount) Synopsis() string { return "Mount a file system for access on your machine." }

// Usage shows usage
func (cmd *FileSystemMount) Usage() string { return "nerd fs mount FILE_SYSTEM_NAME" }
