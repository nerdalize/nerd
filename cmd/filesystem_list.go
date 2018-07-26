package cmd

import (
    "context"
    "sort"

    humanize "github.com/dustin/go-humanize"
    flags "github.com/jessevdk/go-flags"
    "github.com/mitchellh/cli"
    
    "github.com/nerdalize/nerd/svc"
)

//FileSystemList command
type FileSystemList struct {
    *command
}

//FileSystemListFactory creates the command
func FileSystemListFactory(ui cli.Ui) cli.CommandFactory {
    cmd := &FileSystemList{}
    cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, nil, flags.None, "nerd fs list")
    return func() (cli.Command, error) {
        return cmd, nil
    }
}

//Execute runs the command
func (cmd *FileSystemList) Execute(args []string) (err error) {
    if len(args) > 0 {
        return errShowUsage(MessageNoArgumentRequired)
    }
    kopts := cmd.globalOpts.KubeOpts
    deps, err := NewDeps(cmd.Logger(), kopts)
    if err != nil {
        return renderConfigError(err, "failed to configure")
    }

    ctx := context.Background()
    ctx, cancel := context.WithTimeout(ctx, kopts.Timeout)
    defer cancel()

    in := &svc.ListFileSystemsInput{}
    kube := svc.NewKube(deps)
    out, err := kube.ListFileSystems(ctx, in)
    if err != nil {
        return renderServiceError(err, "failed to list file systems")
    }

    if len(out.Items) == 0 {
        cmd.out.Infof("No file system found.")
        return nil
    }

    sort.Slice(out.Items, func(i int, j int) bool {
        return out.Items[i].Details.CreatedAt.After(out.Items[j].Details.CreatedAt)
    })

    hdr := []string{"FILE SYSTEM", "CREATED AT", "SIZE", "STATUS"}
    rows := [][]string{}
    for _, item := range out.Items {
        rows = append(rows, []string{
            item.Name,
            humanize.Time(item.Details.CreatedAt),
            humanize.Bytes(item.Details.Size),
            string(item.Details.Status),
        })
    }

    return cmd.out.Table(hdr, rows)
}

// Description returns long-form help text
func (cmd *FileSystemList) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *FileSystemList) Synopsis() string { return "Return file systems that are managed by the cluster." }

// Usage shows usage
func (cmd *FileSystemList) Usage() string { return "nerd fs list [OPTIONS]" }
