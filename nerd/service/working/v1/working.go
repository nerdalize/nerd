package v1working

import (
	"context"
	"log"
	"os/exec"
	"time"

	"github.com/nerdalize/nerd/nerd/client/batch/v1"
	"github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//Worker is a longer running process that spawns processes based on task runs that arrive via the batch client
type Worker struct {
	conf    Conf
	bclient workerClient
	logs    *log.Logger
	qops    v1batch.QueueOps
	pid     string
	qid     string
}

type workerClient interface {
	v1batch.ClientTaskInterface
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
func NewWorker(logger *log.Logger, bclient workerClient, qops v1batch.QueueOps, projectID string, queueID string, conf *Conf) (w *Worker) {
	w = &Worker{
		conf:    *conf,
		logs:    logger,
		bclient: bclient,
		qops:    qops,
		qid:     queueID,
		pid:     projectID,
	}

	return w
}

type runReceive struct {
	run *v1payload.Run
	err error
}

func (w *Worker) startRunExecHeartbeat(procCtx context.Context, procCancel context.CancelFunc, run *v1payload.Run) {
	w.logs.Printf("[DEBUG] Start task run heartbeat")
	defer w.logs.Printf("[DEBUG] Exited task run heartbeat")

	ticker := time.Tick(w.conf.HeartbeatInterval)
	for {
		select {
		case <-procCtx.Done():
			return
		case <-ticker:
			//@TODO send heartbeat
			//@TODO on expire, cancel proc context
		}
	}
}

func (w *Worker) startRunExec(ctx context.Context, run *v1payload.Run) {
	w.logs.Printf("[DEBUG] Start task run execution")
	defer w.logs.Printf("[DEBUG] Exited task run execution")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel() //cancel heartbeat context if this function exits

	cmd := exec.CommandContext(ctx, "false")
	err := cmd.Start()
	if err != nil {
		w.logs.Printf("[ERROR] failed to start run process: %+v", err)
		return
	}

	//start heartbeat, it may use the cancel function to kill the process
	go w.startRunExecHeartbeat(ctx, cancel, run)

	//wait for process to exit
	err = cmd.Wait()
	if err != nil {
		switch e := err.(type) {
		case *exec.ExitError:
			w.logs.Printf("[INFO] run process exited: %v", e)
			//@TODO store a piece of stderr
			//@TODO includes process state
		default:
			w.logs.Printf("[ERROR] run process exited unexpectedly: %+v", err)
		}
	} else {
		w.logs.Printf("[INFO] run process exited succesfully")
		//@TODO send task success
	}
}

func (w *Worker) startReceivingRuns(ctx context.Context) <-chan runReceive {
	runCh := make(chan runReceive)
	go func() {
		w.logs.Printf("[DEBUG] Start receiving task runs")
		defer w.logs.Printf("[DEBUG] Exited task run receiving")

		for {
			select {
			case <-ctx.Done():
				return
			default:

				//@TODO we should allow context to be passed on to the batch client to allow cancelling of tcp connections
				out, err := w.bclient.ReceiveTaskRuns(w.pid, w.qid, w.conf.ReceiveTimeout, w.qops)
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
	w.logs.Printf("[DEBUG] started worker")
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
