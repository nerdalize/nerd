package svc

import (
	"context"
)

//RunJobInput is the input to RunJob
type RunJobInput struct{}

//RunJobOutput is the output to RunJob
type RunJobOutput struct{}

//RunJob will create a job on kubernetes
func (k *Kube) RunJob(ctx context.Context, in *RunJobInput) (out *RunJobOutput, err error) {
	if in == nil || ctx == nil {
		return nil, errNoInput{}
	}

	//@TODO try to use ctx (?)
	//@TODO call the Kubernetes

	return out, nil
}
