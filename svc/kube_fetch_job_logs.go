package svc

import (
	"bytes"
	"context"

	"github.com/nerdalize/nerd/pkg/kubevisor"
	corev1 "k8s.io/api/core/v1"
)

//FetchJobLogsInput is the input to FetchJobLogs
type FetchJobLogsInput struct {
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

	pods := &pods{}
	err = k.visor.ListResources(ctx, kubevisor.ResourceTypePods, pods, nil)
	if err != nil {
		return nil, err
	}

	if len(pods.Items) < 1 {
		return nil, errNoLogs{reasonNoPods: true}
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
	err = k.visor.FetchLogs(ctx, 100, buf, "main", last.GetName())
	if err != nil {
		return nil, err //@TODO, possible race, at this point the pod could have been deleted
	}

	return &FetchJobLogsOutput{Data: buf.Bytes()}, nil
}
