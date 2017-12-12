package svc

import (
	"context"
)

//FetchJobLogsInput is the input to FetchJobLogs
type FetchJobLogsInput struct {
	Name string `validate:"min=1,printascii"`
}

//FetchJobLogsOutput is the output to FetchJobLogs
type FetchJobLogsOutput struct{}

//FetchJobLogs will create a job on kubernetes
func (k *Kube) FetchJobLogs(ctx context.Context, in *FetchJobLogsInput) (out *FetchJobLogsOutput, err error) {
	if err = k.checkInput(ctx, in); err != nil {
		return nil, err
	}

	//@TODO get all pods with prefix and label

	//@TODO get pod name
	//@TODO assume container name

	return &FetchJobLogsOutput{}, nil
}
