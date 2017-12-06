package cmd

import (
	"context"

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

	hdr := []string{"JOB", "IMAGE", "STATUS", "PHASE", "WAITING REASON"}
	rows := [][]string{}
	for _, item := range out.Items {

		//@TODO think of a better way to format jobs
		status := "Unkown"
		if item.DeletedAt.IsZero() {
			if !item.FailedAt.IsZero() {
				status = "Failed"
			} else {
				if !item.CompletedAt.IsZero() {
					status = "Completed"
				} else {

					if !item.ActiveAt.IsZero() {
						status = "Active" //@TODO at this point the job's sole pod can still be:
						// - Pending (due to capacity, not being placed)
						// - ErrImagePull (due to wrong image being provided)
						// - Running (successfully being in progress)
					}
				}
			}
		} else {
			status = "Deleting..."
		}

		rows = append(rows, []string{item.Name, item.Image, status, string(item.Details.Phase), item.Details.WaitingReason})
	}

	return cmd.out.Table(hdr, rows)
}

// Description returns long-form help text
func (cmd *JobList) Description() string { return PlaceholderHelp }

// Synopsis returns a one-line
func (cmd *JobList) Synopsis() string { return PlaceholderSynopsis }

// Usage shows usage
func (cmd *JobList) Usage() string { return PlaceholderUsage }
