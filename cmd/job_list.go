package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	humanize "github.com/dustin/go-humanize"
	flags "github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/svc"
)

//JobList command
type JobList struct {
	*command
}

//JobListFactory creates the command
func JobListFactory(ui cli.Ui) cli.CommandFactory {
	cmd := &JobList{}
	cmd.command = createCommand(ui, cmd.Execute, cmd.Description, cmd.Usage, cmd, nil, flags.None, "nerd job list")
	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//Execute runs the command
func (cmd *JobList) Execute(args []string) (err error) {
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

	kube := svc.NewKube(deps)

	qin := &svc.ListQuotasInput{}
	qout, err := kube.ListQuotas(ctx, qin)
	if err != nil {
		return renderServiceError(err, "failed to list quotas")
	}

	in := &svc.ListJobsInput{}

	out, err := kube.ListJobs(ctx, in)
	if err != nil {
		return renderServiceError(err, "failed to list jobs")
	}

	if len(out.Items) == 0 {
		cmd.out.Infof("No job found.")
		return nil
	}

	cmd.out.Infof("All your jobs are listed below. To see the logs of a specific job, you can use: `nerd job logs <JOB-NAME>`")
	var q *svc.ListQuotaItem
	if len(qout.Items) > 0 {
		q = qout.Items[0]
		percMem := 100 * (float64(q.UseRequestMemory) / float64(q.RequestMemory))
		percVCPU := 100 * (float64(q.UseRequestCPU) / float64(q.RequestCPU))
		memrow := []string{"Memory Usage:", fmt.Sprintf("%s / %s GB", renderMemory(q.UseRequestMemory), renderMemory(q.RequestMemory)), fmt.Sprintf("(%.1f%%)", percMem)}
		vcpurow := []string{"vCPU Usage:", fmt.Sprintf("%s / %s Core(s)", renderVCPU(q.UseRequestCPU), renderVCPU(q.RequestCPU)), fmt.Sprintf("(%.1f%%)", percVCPU)}

		cmd.out.Table([]string{"", "", "", ""}, [][]string{
			memrow,
			vcpurow,
		})

		cmd.out.Info("")
	}

	sort.Slice(out.Items, func(i int, j int) bool {
		return out.Items[i].CreatedAt.After(out.Items[j].CreatedAt)
	})

	hdr := []string{
		"JOB",
		"IMAGE",
		"INPUT",
		"OUTPUT",
		"MEMORY",
		"VCPU",
		"CREATED AT",
		"PHASE",
		"DETAILS",
	}

	rows := [][]string{}
	for _, item := range out.Items {
		rows = append(rows, []string{
			item.Name,
			item.Image,
			strings.Join(item.Input, ","),
			strings.Join(item.Output, ","),
			renderMemory(item.Memory),
			renderVCPU(item.VCPU),
			humanize.Time(item.CreatedAt),
			renderItemPhase(item),
			strings.Join(renderItemDetails(item, q), ","),
		})
	}

	return cmd.out.Table(hdr, rows)
}

// Description returns long-form help text
func (cmd *JobList) Description() string { return cmd.Synopsis() }

// Synopsis returns a one-line
func (cmd *JobList) Synopsis() string { return "Return jobs that are managed by the cluster" }

// Usage shows usage
func (cmd *JobList) Usage() string { return "nerd job list [OPTIONS]" }

func renderMemory(n int64) string {
	return fmt.Sprintf("%.1f", float64(n/1000/1000/1000)/1000)
}

func renderVCPU(n int64) string {
	return fmt.Sprintf("%.1f", float64(n)/1000)
}

func renderItemDetails(item *svc.ListJobItem, quota *svc.ListQuotaItem) (details []string) {
	if item.Details.TerminatedExitCode != 0 {
		details = append(details, fmt.Sprintf("Non-zero exit code: %d", item.Details.TerminatedExitCode))
	}

	if item.Details.WaitingReason != "" {
		wreason := item.Details.WaitingReason
		if strings.Contains(wreason, "Image") {
			wreason = "Failure while pulling image"
		}

		if wreason == "ContainerCreating" {
			wreason = "Creating Container"
		}

		details = append(details, wreason)
	}

	if item.Details.UnschedulableReason != "" {
		usreason := item.Details.UnschedulableReason
		if strings.Contains(usreason, "NotYetSchedulable") {
			//this is only shown when there is no quota configured
			//@TODO what if the quota is large enough but a job will never fit on a node?
			usreason = "Not enough cluster resources"
		}

		details = append(details, usreason)
	}

	if len(item.Details.FailedCreateEvents) > 0 {
		lastMsg := ""
		for _, ev := range item.Details.FailedCreateEvents {
			if strings.Contains(ev.Message, "exceeded quota") && quota != nil {
				if item.Memory > quota.RequestMemory || item.VCPU > quota.RequestCPU {
					//user has specified something that will never fit with the curren quota settings
					lastMsg = "Specified resource request exceeds maximum"
				} else if !item.Details.Scheduled {
					//we only show this event if the pod is no scheduled since events stay behind
					//after the situation was fixed in that case don't want to show it anymore
					lastMsg = "Queued for resources"
				}
			}
		}

		if lastMsg != "" {
			details = append(details, lastMsg)
		}
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

	return "Unknown" //by default the status is unknown
}
