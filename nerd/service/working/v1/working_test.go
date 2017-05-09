package v1working_test

import (
	"context"
	"log"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/nerdalize/nerd/nerd/client/batch/v1"
	"github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
	"github.com/nerdalize/nerd/nerd/service/working/v1"
)

//mock queueops
type mQueueOps struct{}

func (mq *mQueueOps) ReceiveMessages(queueURL string, maxNoOfMessages, waitTimeSeconds int64) (messages []interface{}, err error) {
	return messages, nil
}
func (mq *mQueueOps) UnmarshalMessage(message interface{}, v interface{}) error { return nil }
func (mq *mQueueOps) DeleteMessage(queueURL string, message interface{}) error {
	return nil
}

//mock the sdk client
type mClient struct {
	receiveRuns chan []*v1payload.Run
	receiveErrs chan error
}

func (c *mClient) StartTask(projectID, queueID, payload string) (output *v1payload.StartTaskOutput, err error) {
	return output, nil
}
func (c *mClient) StopTask(projectID, queueID string, taskID int64) (output *v1payload.StopTaskOutput, err error) {
	return output, nil
}
func (c *mClient) ListTasks(projectID, queueID string) (output *v1payload.ListTasksOutput, err error) {
	return output, nil
}
func (c *mClient) DescribeTask(projectID, queueID string, taskID int64) (output *v1payload.DescribeTaskOutput, err error) {
	return output, nil
}
func (c *mClient) ReceiveTaskRuns(projectID, queueID string, timeout time.Duration, queueOps v1batch.QueueOps) (output []*v1payload.Run, err error) {
	return <-c.receiveRuns, <-c.receiveErrs
}
func (c *mClient) SendRunHeartbeat(projectID, queueID string, taskID int64, runToken string) (output *v1payload.SendRunHeartbeatOutput, err error) {
	return output, nil
}
func (c *mClient) SendRunSuccess(projectID, queueID string, taskID int64, runToken, result string) (output *v1payload.SendRunSuccessOutput, err error) {
	return output, nil
}
func (c *mClient) SendRunFailure(projectID, queueID string, taskID int64, runToken, errCode, errMessage string) (output *v1payload.SendRunFailureOutput, err error) {
	return output, nil
}

func TestContextDone(t *testing.T) {
	logs := log.New(os.Stderr, "test/", log.Lshortfile)

	bclient := &mClient{}
	qops := &mQueueOps{}

	w := v1working.NewWorker(logs, bclient, qops, "project-x", "queue-y", v1working.DefaultConf())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	w.Start(ctx)

	if runtime.NumGoroutine() != 2 {
		t.Fatalf("expected 2 goroutines, got: %v", runtime.NumGoroutine())
	}
}

func TestRunReceiving(t *testing.T) {
	logs := log.New(os.Stderr, "test/", log.Lshortfile)

	bclient := &mClient{
		receiveRuns: make(chan []*v1payload.Run, 1),
		receiveErrs: make(chan error, 1)}

	bclient.receiveErrs <- nil
	bclient.receiveRuns <- []*v1payload.Run{
		{ProjectID: "my-project", TaskID: 123, QueueID: "my-queue", Token: "my-token"},
	}

	qops := &mQueueOps{}

	w := v1working.NewWorker(logs, bclient, qops, "project-x", "queue-y", v1working.DefaultConf())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	w.Start(ctx)

	time.Sleep(time.Second) //@TODO wait for everything to close down properly

	//@TODO assert something
}
