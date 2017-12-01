package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"

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
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{},
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

	err = k.visor.CreateResource(ctx, kubevisor.KubeResourceTypeJobs, job, in.Name)
	if err != nil {
		return nil, err
	}

	return &RunJobOutput{
		Name: job.Name,
	}, nil
}
