package svc

import (
	"context"
	"encoding/hex"

	"github.com/nerdalize/nerd/pkg/kubevisor"
	"github.com/pkg/errors"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	//JobDefaultBackoffLimit determines how often we will retry a pod's job on when its failing
	JobDefaultBackoffLimit = int32(3)
)

//RunJobInput is the input to RunJob
type RunJobInput struct {
	Image                string `validate:"min=1"`
	Name                 string `validate:"printascii"`
	Env                  map[string]string
	BackoffLimit         *int32
	Args                 []string
	Volumes              []JobVolume
	FileSystemMounts	 []FileSystemMount
	Memory               string
	VCPU                 string
	Secret               string
}

//JobVolumeType determines if its content will be uploaded or downloaded
type JobVolumeType string

const (
	//JobVolumeTypeInput determines the volume to be input
	JobVolumeTypeInput = JobVolumeType("input")

	//JobVolumeTypeOutput determines the volume to output
	JobVolumeTypeOutput = JobVolumeType("output")
)

//JobVolume can be used in a job
type JobVolume struct {
	MountPath     string `validate:"is-abs-path"`
	InputDataset  string
	OutputDataset string
}

type FileSystemMount struct {
	FileSystemName string `validate:"printascii`
	MountPath      string `validate:"is-abs-path"`
	SubPath		   string `validate:"is-abs-path"`
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
							Name:            "main",
							Image:           in.Image,
							ImagePullPolicy: "Always",
							Env:             envs,
							Args:            in.Args,
						},
					},
				},
			},
		},
	}
	if in.Memory != "" || in.VCPU != "" {
		resources, err := getResources(in.Memory, in.VCPU)
		if err != nil {
			return nil, err
		}
		job.Spec.Template.Spec.Containers[0].Resources = resources
	}

	for _, vol := range in.Volumes {
		opts := map[string]string{}
		if vol.InputDataset != "" {
			opts["input/dataset"] = vol.InputDataset
		}

		if vol.OutputDataset != "" {
			opts["output/dataset"] = vol.OutputDataset
		}

		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, v1.Volume{
			Name: hex.EncodeToString([]byte(vol.MountPath)),
			VolumeSource: v1.VolumeSource{
				FlexVolume: &v1.FlexVolumeSource{
					Driver:  "nerdalize.com/dataset",
					Options: opts,
				},
			},
		})

		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name:      hex.EncodeToString([]byte(vol.MountPath)),
			MountPath: vol.MountPath,
		})
	}

	for _, mount := range in.FileSystemMounts {
		job.Spec.Template.Spec.Containers[0].VolumeMounts = append(job.Spec.Template.Spec.Containers[0].VolumeMounts, v1.VolumeMount{
			Name: 	   hex.EncodeToString([]byte(mount.MountPath)),
			MountPath: mount.MountPath,
			SubPath:   mount.SubPath,
		})

		job.Spec.Template.Spec.Volumes = append(job.Spec.Template.Spec.Volumes, v1.Volume{
			Name: 		  hex.EncodeToString([]byte(mount.MountPath)),
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: mount.FileSystemName,
				},
			},
		})
	}

	if in.Secret != "" {
		job.Spec.Template.Spec.ImagePullSecrets = append(job.Spec.Template.Spec.ImagePullSecrets, v1.LocalObjectReference{
			Name: kubevisor.DefaultPrefix + in.Secret,
		})
	}

	err = k.visor.CreateResource(ctx, kubevisor.ResourceTypeJobs, job, in.Name)
	if err != nil {
		return nil, err
	}

	return &RunJobOutput{
		Name: job.Name,
	}, nil
}

func getResources(memory, vcpu string) (v1.ResourceRequirements, error) {
	m, err := resource.ParseQuantity(memory)
	if err != nil {
		return v1.ResourceRequirements{}, errors.Wrap(err, "could not create memory resource")
	}

	cpu, err := resource.ParseQuantity(vcpu)
	if err != nil {
		return v1.ResourceRequirements{}, errors.Wrap(err, "could not create cpu resource")
	}

	return v1.ResourceRequirements{
		Limits:   v1.ResourceList{v1.ResourceCPU: cpu, v1.ResourceMemory: m},
		Requests: v1.ResourceList{v1.ResourceCPU: cpu, v1.ResourceMemory: m},
	}, nil
}
