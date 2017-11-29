package svc

import (
	"context"

	"github.com/pkg/errors"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//RunJobInput is the input to RunJob
type RunJobInput struct {
	Image string `validate:"min=1"`
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

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "j-",
		},
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

	job, err = k.api.BatchV1().Jobs(k.ns).Create(job)
	if err != nil {
		//@TODO add our own error
		return nil, errors.Wrap(err, "failed to create job")
	}

	//@TODO try to use ctx (?)
	//@TODO call the Kubernetes
	out = &RunJobOutput{Name: job.Name}

	return out, nil
}
