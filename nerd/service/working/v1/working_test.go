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

type mRunFeedback struct {
	projectID  string
	workloadID string
	taskID     int64
	token      string
	result     string
	errCode    string
	errMsg     string
}

//mock the sdk client
type mClient struct {
	receiveRuns   chan []*v1payload.Run
	receiveErrs   chan error
	runHeartbeats chan mRunFeedback
	runSuccess    chan mRunFeedback
	runFailure    chan mRunFeedback
}

func (c *mClient) StartTask(projectID, workloadID string, cmd []string, env map[string]string, stdin []byte) (output *v1payload.StartTaskOutput, err error) {
	return output, nil
}
func (c *mClient) StopTask(projectID, workloadID string, taskID int64) (output *v1payload.StopTaskOutput, err error) {
	return output, nil
}
func (c *mClient) ListTasks(projectID, workloadID string, onlySuccessTasks bool) (output *v1payload.ListTasksOutput, err error) {
	return output, nil
}
func (c *mClient) DescribeTask(projectID, workloadID string, taskID int64) (output *v1payload.DescribeTaskOutput, err error) {
	return output, nil
}
func (c *mClient) PatchTask(projectID, workloadID string, taskID int64, outputDatasetID string) (output *v1payload.PatchTaskOutput, err error) {
	return output, nil
}
func (c *mClient) ReceiveTaskRuns(projectID, workloadID string, timeout time.Duration, queueOps v1batch.QueueOps) (output []*v1payload.Run, err error) {
	return <-c.receiveRuns, <-c.receiveErrs
}
func (c *mClient) SendRunHeartbeat(projectID, workloadID string, taskID int64, runToken string) (output *v1payload.SendRunHeartbeatOutput, err error) {
	c.runHeartbeats <- mRunFeedback{projectID: projectID, workloadID: workloadID, taskID: taskID, token: runToken}
	return output, nil
}
func (c *mClient) SendRunSuccess(projectID, workloadID string, taskID int64, runToken, result string) (output *v1payload.SendRunSuccessOutput, err error) {
	c.runSuccess <- mRunFeedback{projectID: projectID, workloadID: workloadID, taskID: taskID, token: runToken, result: result}
	return output, nil
}
func (c *mClient) SendRunFailure(projectID, workloadID string, taskID int64, runToken, errCode, errMessage string) (output *v1payload.SendRunFailureOutput, err error) {
	c.runFailure <- mRunFeedback{projectID: projectID, workloadID: workloadID, taskID: taskID, token: runToken, errCode: errCode, errMsg: errMessage}
	return output, nil
}

func TestContextDone(t *testing.T) {
	logs := log.New(os.Stderr, "test/", log.Lshortfile)

	bclient := &mClient{}
	qops := &mQueueOps{}

	w := v1working.NewWorker(logs, bclient, qops, "project-x", "workload-y", "false", []string{}, nil, v1working.DefaultConf())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	w.Start(ctx)

	if runtime.NumGoroutine() != 4 {
		t.Fatalf("expected 4 goroutines, got: %v", runtime.NumGoroutine())
	}
}

func TestRunReceivingFailingTask(t *testing.T) {
	logs := log.New(os.Stderr, "test/", log.Lshortfile)

	bclient := &mClient{
		receiveRuns:   make(chan []*v1payload.Run, 1),
		receiveErrs:   make(chan error, 1),
		runHeartbeats: make(chan mRunFeedback, 1),
		runSuccess:    make(chan mRunFeedback, 1),
		runFailure:    make(chan mRunFeedback, 1),
	}

	bclient.receiveErrs <- nil
	bclient.receiveRuns <- []*v1payload.Run{
		{ProjectID: "my-project", TaskID: 123, WorkloadID: "my-workload", Token: "my-token"},
	}

	qops := &mQueueOps{}

	w := v1working.NewWorker(logs, bclient, qops, "project-x", "workload-y", "false", []string{}, nil, v1working.DefaultConf())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	w.Start(ctx)

	failure := <-bclient.runFailure
	if failure.errCode == v1working.RunErrCodeUnexpected {
		t.Fatalf("expected failure not to be unexpected")
	}

	if failure.errMsg != "" {
		t.Fatalf("error message to be empty, got: %v", failure.errMsg)
	}
}

func TestRunReceivingSuccessTask(t *testing.T) {
	logs := log.New(os.Stderr, "test/", log.Lshortfile)

	bclient := &mClient{
		receiveRuns:   make(chan []*v1payload.Run, 1),
		receiveErrs:   make(chan error, 1),
		runHeartbeats: make(chan mRunFeedback, 1),
		runSuccess:    make(chan mRunFeedback, 1),
		runFailure:    make(chan mRunFeedback, 1),
	}

	bclient.receiveErrs <- nil
	bclient.receiveRuns <- []*v1payload.Run{
		{ProjectID: "my-project", TaskID: 123, WorkloadID: "my-workload", Token: "my-token", Cmd: []string{"2"}},
	}

	qops := &mQueueOps{}
	conf := v1working.DefaultConf()
	conf.HeartbeatInterval = time.Second

	w := v1working.NewWorker(logs, bclient, qops, "project-x", "workload-y", "sleep", []string{}, nil, conf)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	go w.Start(ctx)
	defer cancel()

	heartbeat := <-bclient.runHeartbeats //after a second (as configured)
	if heartbeat.token != "my-token" {
		t.Fatalf("expected token send to receive")
	}

	success := <-bclient.runSuccess //after 2 seconds as send in the task
	if success.errCode != "" {
		t.Fatalf("expected error code to be empty, got: %v", success.errCode)
	}

	if success.errMsg != "" {
		t.Fatalf("error message to be empty, got: %v", success.errMsg)
	}

	if success.result != "exit status 0" {
		t.Fatalf("expected success result to be exit status 0")
	}
}
