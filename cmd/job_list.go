package cmd

import (
	"context"
	"sort"
	"strings"

	humanize "github.com/dustin/go-humanize"
	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
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
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, flags.None)
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *JobList) Execute(args []string) (err error) {
	kopts := cmd.KubeOpts
	deps, err := NewDeps(cmd.Logger(), kopts)
	if err != nil {
		return renderConfigError(err, "failed to configure")
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, cmd.Timeout)
	defer cancel()

	in := &svc.ListJobsInput{}
	kube := svc.NewKube(deps)
	out, err := kube.ListJobs(ctx, in)
	if err != nil {
		return renderServiceError(err, "failed to list jobs")
	}

	cmd.out.Infof("To see the logs of a job, use: `nerd job logs <JOB-NAME>`")

	sort.Slice(out.Items, func(i int, j int) bool {
		return out.Items[i].CreatedAt.After(out.Items[j].CreatedAt)
	})
	hdr := []string{"JOB", "IMAGE", "INPUT", "OUTPUT", "CREATED AT", "PHASE", "DETAILS"}
	rows := [][]string{}
	for _, item := range out.Items {
		rows = append(rows, []string{
			item.Name,
			item.Image,
			item.Input,
			item.Output,
			humanize.Time(item.CreatedAt),
			renderItemPhase(item),
			strings.Join(renderItemDetails(item), ","),
		})
	}

	return cmd.out.Table(hdr, rows)
}

// Description returns long-form help text
func (cmd *JobList) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *JobList) Synopsis() string { return "Return jobs that are managed by the cluster" }

// Usage shows usage
func (cmd *JobList) Usage() string { return "nerd job list" }

func renderItemDetails(item *svc.ListJobItem) (details []string) {
	if item.Details.WaitingReason != "" {
		wreason := item.Details.WaitingReason
		if strings.Contains(wreason, "Image") {
			wreason = "Failure while pulling image"
		}

		details = append(details, wreason)
	}

	if item.Details.UnschedulableReason != "" {
		usreason := item.Details.UnschedulableReason
		if strings.Contains(usreason, "NotYetSchedulable") {
			usreason = "No resources available"
		}

		details = append(details, usreason)
	}

	return details
}

func renderItemPhase(item *svc.ListJobItem) string {
	if !item.DeletedAt.IsZero() {
		return "Deleting" //in progress of deleting
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

	if !item.Details.Scheduled {
		return "Waiting" //waiting to be scheduled
	}

	if !item.ActiveAt.IsZero() {
		if item.Details.Phase != "" {
			if item.Details.Phase == svc.JobDetailsPhasePending {
				return "Starting" //if not "waiting" but pending, call it "starting" instead
			}

			return string(item.Details.Phase) //detailed phase is always more usefull then active
		}

		return "Active" //little to go on, but better then nothing
	}

	return "Unkown" //by default the status is unkown
}
