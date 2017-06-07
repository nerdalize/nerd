package v1working

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/nerdalize/nerd/nerd/client/batch/v1"
	"github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
	v1datatransfer "github.com/nerdalize/nerd/nerd/service/datatransfer/v1"
)

var (
	//RunErrCodeUnexpected is presented to the api when an unexpected error occured during the run of the task
	RunErrCodeUnexpected = "ERR_UNEXPECTED"

	//RunResultUndefined is send to the server when
	RunResultUndefined = "-"
)

//Worker is a longer running process that spawns processes based on task runs that arrive via the batch client
type Worker struct {
	conf       Conf
	batch      workerClient
	logs       *log.Logger
	qops       v1batch.QueueOps
	pid        string
	wid        string
	uploadConf *v1datatransfer.UploadConfig

	entrypoint []string
	cmd        []string
	// bexec string
	// bargs []string
}

type workerClient interface {
	v1batch.ClientTaskInterface
	v1batch.ClientRunInterface
}

//Conf holds worker configuration
type Conf struct {
	ReceiveTimeout    time.Duration
	HeartbeatInterval time.Duration
}

//DefaultConf creates a sensible configuration
func DefaultConf() *Conf {
	return &Conf{
		ReceiveTimeout:    time.Second * 20,
		HeartbeatInterval: time.Second * 20,
	}
}

//NewWorker creates a worker based on the provided configuration
func NewWorker(logger *log.Logger, batchClient workerClient, qops v1batch.QueueOps, projectID string, workloadID string, entrypoint, cmd []string, uploadConf *v1datatransfer.UploadConfig, conf *Conf) (w *Worker) {
	w = &Worker{
		conf:       *conf,
		logs:       logger,
		batch:      batchClient,
		qops:       qops,
		wid:        workloadID,
		pid:        projectID,
		uploadConf: uploadConf,

		entrypoint: entrypoint,
		cmd:        cmd,
		// bexec: baseExec,
		// bargs: baseArgs,
	}

	return w
}

type runReceive struct {
	run *v1payload.Run
	err error
}

func (w *Worker) startRunExecHeartbeat(procCtx context.Context, cancelProc context.CancelFunc, run *v1payload.Run) {
	w.logs.Printf("[DEBUG] Start task run heartbeat")
	defer w.logs.Printf("[DEBUG] Exited task run heartbeat")

	ticker := time.Tick(w.conf.HeartbeatInterval)
	for {
		select {
		case <-procCtx.Done():
			return
		case <-ticker:
			if out, err := w.batch.SendRunHeartbeat(run.ProjectID, run.WorkloadID, run.TaskID, run.Token); err != nil {
				w.logs.Printf("[ERROR] failed to send run heartbeat: %v", err)
			} else if out != nil && out.HasExpired {
				cancelProc()
			}

		}
	}
}

func (w *Worker) startRunExec(ctx context.Context, run *v1payload.Run) {
	w.logs.Printf("[DEBUG] Start task run execution")
	defer w.logs.Printf("[DEBUG] Exited task run execution")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel() //cancel heartbeat context if this function exits

	//blindly appending task args to base args
	command := w.cmd
	if len(run.Cmd) > 0 {
		command = run.Cmd
	}
	command = append(w.entrypoint, command...)
	if len(command) == 0 {
		w.logs.Print("[ERROR] no run command was specified")
		return
	}
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Stdin = bytes.NewBuffer(run.Stdin)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	for k, v := range run.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	err := cmd.Start()
	if err == nil {

		//start heartbeat, it may use the cancel() to kill the process
		go w.startRunExecHeartbeat(ctx, cancel, run)

		//we launched successfully, set err to process result
		err = cmd.Wait()
	}

	//if an error happend at this point we want to send a failure to the server
	if err != nil {
		var errCode string
		var errMsg string
		switch e := err.(type) {
		case *exec.ExitError:
			errCode = e.String()
			errMsg = base64.StdEncoding.EncodeToString(e.Stderr)
		default:
			w.logs.Printf("[ERROR] run process exited unexpectedly: %+v", err)
			errCode = RunErrCodeUnexpected
			errMsg = e.Error()
		}

		//@TODO allow sending context
		if _, err = w.batch.SendRunFailure(
			run.ProjectID,
			run.WorkloadID,
			run.TaskID,
			run.Token,
			errCode,
			errMsg,
		); err != nil {
			w.logs.Printf("[ERROR] failed to send run failure: %+v", err)
		}

	} else {
		runRes := RunResultUndefined
		if cmd.ProcessState != nil {
			runRes = cmd.ProcessState.String()
		}

		w.logs.Printf("[INFO] run process exited succesfully")
		outputDatasetID := ""
		if w.uploadConf != nil {
			var empty bool
			if empty, err = IsEmptyDir(w.uploadConf.LocalDir); !empty && err == nil {
				w.logs.Printf("[INFO] uploading output data")
				var ds *v1payload.DatasetSummary
				ds, err = v1datatransfer.Upload(ctx, *w.uploadConf)
				if err != nil {
					w.logs.Printf("[ERROR] failed to upload output dataset: %+v", err)
					if _, err = w.batch.SendRunFailure(
						run.ProjectID,
						run.WorkloadID,
						run.TaskID,
						run.Token,
						"0",
						"failed to output upload dataset",
					); err != nil {
						w.logs.Printf("[ERROR] failed to send run failure: %+v", err)
					}
					return
				}
				outputDatasetID = ds.DatasetID
				if err = RemoveContents(w.uploadConf.LocalDir); err != nil {
					w.logs.Printf("[ERROR] failed to clear output directory '%v', shutting down", err)
					cancel()
				}
			}
		}
		//@TODO allow sending context
		if _, err = w.batch.SendRunSuccess(
			run.ProjectID,
			run.WorkloadID,
			run.TaskID,
			run.Token,
			runRes,
			outputDatasetID,
		); err != nil {
			w.logs.Printf("[ERROR] failed to send run success: %+v", err)
		}
	}
}

func (w *Worker) startReceivingRuns(ctx context.Context) <-chan runReceive {
	runCh := make(chan runReceive)
	go func() {
		w.logs.Printf("[DEBUG] Started receiving task runs, base exec: '%v %v'", w.entrypoint, w.cmd)
		defer w.logs.Printf("[DEBUG] Exited task run receiving")

		for {
			select {
			case <-ctx.Done():
				return
			default:

				//@TODO we should allow context to be passed on to the batch client to allow cancelling of tcp connections
				out, err := w.batch.ReceiveTaskRuns(w.pid, w.wid, w.conf.ReceiveTimeout, w.qops)
				if err != nil {
					runCh <- runReceive{err: err}
					continue
				}

				for _, run := range out {
					runCh <- runReceive{run: run}
				}
			}
		}
	}()

	return runCh
}

//Start will block and begins handling tasks run. It stops when context ctx ends
func (w *Worker) Start(ctx context.Context) {
	w.logs.Printf("[DEBUG] started worker for workload '%s'", w.wid)
	defer w.logs.Printf("[DEBUG] exited worker")

	runCh := w.startReceivingRuns(ctx)
	for {
		select {
		case r := <-runCh:
			if r.err != nil {
				w.logs.Printf("[ERROR] Failed to receive task run: %v", r.err)
				break
			}

			w.logs.Printf("[INFO] Received run: %#v", r.run)
			go w.startRunExec(ctx, r.run)

		case <-ctx.Done():
			return
		}
	}
}
