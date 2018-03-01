package svc

import (
	"bytes"
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

//FetchJobLogsInput is the input to FetchJobLogs
type FetchJobLogsInput struct {
	Tail int64  `validate:"min=0"`
	Name string `validate:"min=1,printascii"`
}

//FetchJobLogsOutput is the output to FetchJobLogs
type FetchJobLogsOutput struct {
	Data []byte
}

//FetchJobLogs will create a job on kubernetes
func (k *Kube) FetchJobLogs(ctx context.Context, in *FetchJobLogsInput) (out *FetchJobLogsOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	job := &batchv1.Job{}
	err = k.visor.GetResource(ctx, kubevisor.ResourceTypeJobs, job, in.Name)
	if err != nil {
		return nil, err
	}

	pods := &pods{}
	err = k.visor.ListResources(ctx, kubevisor.ResourceTypePods, pods, []string{"controller-uid=" + string(job.GetUID())}, []string{})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) < 1 {
		return &FetchJobLogsOutput{}, nil
	}

	var last corev1.Pod
	for _, pod := range pods.Items {
		if pod.CreationTimestamp.Local().After(last.CreationTimestamp.Local()) {
			last = pod //found a more recent pod
		} else {
			continue
		}
	}

	buf := bytes.NewBuffer(nil)
	err = k.visor.FetchLogs(ctx, in.Tail, buf, "main", last.GetName())
	if err != nil {
		if kubevisor.IsNotExistsErr(err) {
			return nil, errRaceCondition{err} //pod was deleted since we listed it
		}

		return nil, err
	}

	return &FetchJobLogsOutput{Data: buf.Bytes()}, nil
}
