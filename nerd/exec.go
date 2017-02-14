package nerd

import "github.com/nerdalize/nerd/nerd/payload"

//Executor allows execution of a Task
type Executor interface {
	Execute(*payload.Task) error
}

//DockerExecutor runs tasks using the Docker command line interface
type DockerExecutor struct{}

//Execute a task
func (exec *DockerExecutor) Execute(t *payload.Task) error {
	return ErrNotImplemented
}

//KubeExecutor runs a task by passing it to Kubernetes
type KubeExecutor struct{}

//Execute a task
func (exec *KubeExecutor) Execute(t *payload.Task) error {
	return ErrNotImplemented
}
