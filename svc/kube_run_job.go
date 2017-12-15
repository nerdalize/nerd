package svc

import (
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	//JobDefaultBackoffLimit determines how often we will retry a pod's job on when its failing
	JobDefaultBackoffLimit = int32(3)
)

//RunJobInput is the input to RunJob
type RunJobInput struct {
	Image        string `validate:"min=1"`
	Name         string `validate:"printascii"`
	Env          map[string]string
	BackoffLimit *int32
	Args         []string
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

	if in.BackoffLimit == nil {
		in.BackoffLimit = &JobDefaultBackoffLimit
	}

	envs := []v1.EnvVar{}
	for k, v := range in.Env {
		envs = append(envs, v1.EnvVar{
			Name:  k,
			Value: v,
		})
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{},
		Spec: batchv1.JobSpec{
			BackoffLimit: in.BackoffLimit,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"nerd-app": "cli",
					},
				},
				Spec: v1.PodSpec{
					RestartPolicy: v1.RestartPolicyNever,
					Containers: []v1.Container{
						{
							Name:  "main",
							Image: in.Image,
							Env:   envs,
							Args:  in.Args,
							// Resources: v1.ResourceRequirements{
							// 	Limits: v1.ResourceList{v1.ResourceCPU: resource.MustParse("10"), v1.ResourceMemory: resource.MustParse("256M")},
							// 	// Requests: v1.ResourceList{v1.ResourceCPU: cpu, v1.ResourceMemory: memory},
							// },
						},
					},
				},
			},
		},
	}

	err = k.visor.CreateResource(ctx, kubevisor.ResourceTypeJobs, job, in.Name)
	if err != nil {
		return nil, err
	}

	return &RunJobOutput{
		Name: job.Name,
	}, nil
}
