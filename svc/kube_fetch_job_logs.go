package svc

import (
	"bytes"
	"context"
	"sort"

	"github.com/nerdalize/nerd/pkg/kubevisor"
	batchv1 "k8s.io/api/batch/v1"
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

	//sort by latest created
	sort.Slice(pods.Items, func(i int, j int) bool {
		return pods.Items[i].CreationTimestamp.UnixNano() > pods.Items[j].CreationTimestamp.UnixNano()
	})

	//loop over the pods, return output from the first pod that returns logs, at most 3 times
	buf := bytes.NewBuffer(nil)
	for i := 0; i < len(pods.Items) && i < 3; i++ {
		pod := pods.Items[i]
		_ = k.visor.FetchLogs(ctx, in.Tail, buf, "main", pod.Name)
		if buf.Len() > 0 {
			break
		}
	}

	return &FetchJobLogsOutput{Data: buf.Bytes()}, nil
}
