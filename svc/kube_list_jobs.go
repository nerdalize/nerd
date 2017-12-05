package svc

import (
	"context"
	"time"

	"github.com/nerdalize/nerd/pkg/kubevisor"

	"k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

//ListJobItem is a job listing item
type ListJobItem struct {
	Name        string
	Image       string
	CreatedAt   time.Time
	DeletedAt   time.Time
	ActiveAt    time.Time
	CompletedAt time.Time
	FailedAt    time.Time
}

//ListJobsInput is the input to ListJobs
type ListJobsInput struct{}

//ListJobsOutput is the output to ListJobs
type ListJobsOutput struct {
	Items []*ListJobItem
}

//ListJobs will create a job on kubernetes
func (k *Kube) ListJobs(ctx context.Context, in *ListJobsInput) (out *ListJobsOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	jobs := &jobs{}
	err = k.visor.ListResources(ctx, kubevisor.ResourceTypeJobs, jobs)
	if err != nil {
		return nil, err
	}

	out = &ListJobsOutput{}
	for _, job := range jobs.Items {
		if len(job.Spec.Template.Spec.Containers) != 1 {
			k.logs.Debugf("skipping job '%s' in namespace '%s' as it has not just 1 container", job.Name, job.Namespace)
			continue
		}

		//@TODO if parralism is set to 0, consider "stopped"?

		c := job.Spec.Template.Spec.Containers[0]
		item := &ListJobItem{
			Name:      job.GetName(),
			Image:     c.Image,
			CreatedAt: job.CreationTimestamp.Local(),
		}

		if dt := job.GetDeletionTimestamp(); dt != nil {
			item.DeletedAt = dt.Local() //mark as deleting
		}

		if job.Status.StartTime != nil {
			item.ActiveAt = job.Status.StartTime.Local()
		}

		for _, cond := range job.Status.Conditions {
			if cond.Status != corev1.ConditionTrue {
				continue
			}

			switch cond.Type {
			case v1.JobComplete:
				item.CompletedAt = cond.LastTransitionTime.Local()
			case v1.JobFailed:
				item.FailedAt = cond.LastTransitionTime.Local()
			}
		}

		out.Items = append(out.Items, item)
	}

	//@TODO fetch all pods with label nerd-app=cli and match to jobs

	return out, nil
}

//jobs implements the list transformer interface to allow the kubevisor the manage names for us
type jobs struct{ *v1.JobList }

func (jobs *jobs) Transform(fn func(in kubevisor.ManagedNames) (out kubevisor.ManagedNames)) {
	for i, j1 := range jobs.JobList.Items {
		jobs.Items[i] = *(fn(&j1).(*v1.Job))
	}
}
