package svc

import (
	"context"

	"k8s.io/api/batch/v1"
)

//ListJobItem is a job listing item
type ListJobItem struct {
	Name  string
	Image string
}

//ListJobsInput is the input to ListJobs
type ListJobsInput struct{}

//ListJobsOutput is the output to ListJobs
type ListJobsOutput struct {
	Items []*ListJobItem
}

//ListJobs will create a job on kubernetes
func (k *Kube) ListJobs(ctx context.Context, in *ListJobsInput) (out *ListJobsOutput, err error) {
	jobs := &jobs{}
	err = k.listResource(ctx, KubeResourceTypeJobs, jobs)
	if err != nil {
		return nil, err
	}

	out = &ListJobsOutput{}
	for _, job := range jobs.Items {
		if len(job.Spec.Template.Spec.Containers) != 1 {
			k.logs.Debugf("skipping job '%s' in namespace '%s' as it has not just 1 container", job.Name, job.Namespace)
			continue
		}

		c := job.Spec.Template.Spec.Containers[0]
		out.Items = append(out.Items, &ListJobItem{
			Name:  job.GetName(),
			Image: c.Image,
		})
	}

	return out, nil
}

//jobs implements the list transformer interface
type jobs struct{ *v1.JobList }

func (jobs *jobs) Transform(fn func(in KubeManagedNames) (out KubeManagedNames)) {
	for i, j1 := range jobs.JobList.Items {
		jobs.Items[i] = *(fn(&j1).(*v1.Job))
	}
}
