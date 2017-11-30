package svc

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//RunJobInput is the input to RunJob
type RunJobInput struct {
	Image string `validate:"min=1"`
	Name  string `validate:"printascii"`
}

//RunJobOutput is the output to RunJob
type RunJobOutput struct {
	Name string
}

//RunJob will create a job on kubernetes
func (k *Kube) RunJob(ctx context.Context, in *RunJobInput) (out *RunJobOutput, err error) {
	if in == nil || ctx == nil {
		return nil, errNoInput{}
	}

	err = k.val.StructCtx(ctx, in)
	if err != nil {
		return nil, errValidation{err}
	}

	//@TODO add a nerd prefix

	jobMeta := metav1.ObjectMeta{
		Name: in.Name,
	}

	if jobMeta.Name == "" {
		jobMeta.GenerateName = "j-"
	}

	job := &batchv1.Job{
		ObjectMeta: jobMeta,
		Spec: batchv1.JobSpec{
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					RestartPolicy: v1.RestartPolicyOnFailure,
					Containers: []v1.Container{
						{
							Name:  "main",
							Image: in.Image,
						},
					},
				},
			},
		},
	}

	err = k.CreateResource(ctx, KubeResourceTypeJobs, job)
	if err != nil {
		return nil, err
	}

	return &RunJobOutput{
		Name: job.Name,
	}, nil
}
