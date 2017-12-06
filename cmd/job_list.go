package cmd

import (
	"context"
	"fmt"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
	"github.com/pkg/errors"
)

//JobList command
type JobList struct {
	KubeOpts
	Name string `long:"name" short:"n" description:"assign a name to the job"`

	*command
}

//JobListFactory creates the command
func JobListFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &JobList{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd)
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *JobList) Execute(args []string) (err error) {
	kopts := cmd.KubeOpts
	deps, err := NewDeps(cmd.Logger(), kopts)
	if err != nil {
		return errors.Wrap(err, "failed to configure")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	in := &svc.ListJobsInput{}
	kube := svc.NewKube(deps, kopts.Namespace)
	out, err := kube.ListJobs(ctx, in)
	if err != nil {
		return errors.Wrap(err, "failed to run job")
	}

	hdr := []string{"JOB", "IMAGE", "CREATED AT", "STATUS"}
	rows := [][]string{}
	for _, item := range out.Items {
		details := []string{}
		status := func() string {
			if !item.DeletedAt.IsZero() {
				return "(Deleting)" //in progress of deleting
			}

			if item.Details.Parallelism == 0 {
				return "Stopped"
			}

			if !item.FailedAt.IsZero() {
				return "Failed"
			}

			if !item.CompletedAt.IsZero() {
				return "Completed"
			}

			if !item.ActiveAt.IsZero() {
				if item.Details.Phase != "" {
					return string(item.Details.Phase) //detailed phase is always more usefull then active
				}

				return "Active" //little to go on, but better then nothing
			}

			return "Unkown" //by default the status is unkown
		}()

		//humanize creation
		createdAt := humanize.Time(item.CreatedAt)

		//add explanations as details
		if item.Details.WaitingReason != "" {
			details = append(details, item.Details.WaitingReason)
		}

		if len(details) > 0 {
			status = fmt.Sprintf("%s: %s", status, strings.Join(details, ", "))
		}

		rows = append(rows, []string{item.Name, item.Image, createdAt, status})
	}

	return cmd.out.Table(hdr, rows)
}

// Description returns long-form help text
func (cmd *JobList) Description() string { return PlaceholderHelp }

// Synopsis returns a one-line
func (cmd *JobList) Synopsis() string { return PlaceholderSynopsis }

// Usage shows usage
func (cmd *JobList) Usage() string { return PlaceholderUsage }
