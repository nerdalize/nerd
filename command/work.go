package command

import (
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	naws "github.com/nerdalize/nerd/nerd/aws"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/nerd/payload"
	"github.com/pkg/errors"
)

//WorkOpts describes command options
type WorkOpts struct {
	NerdOpts
}

//Work command
type Work struct {
	*command

	opts   *WorkOpts
	parser *flags.Parser
}

//WorkFactory returns a factory method for the join command
func WorkFactory() func() (cmd cli.Command, err error) {
	cmd := &Work{
		command: &command{
			help:     "",
			synopsis: "start handling tasks on local compute resources",
			parser:   flags.NewNamedParser("nerd work", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &WorkOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//TaskStatus describes the container status of a task
type TaskStatus struct {
	cid   string //container id
	token string //activity token
	code  int    //exit code
	err   error  //application error
	pid   string //project id
	tid   string //task id
}

//DoRun is called by run and allows an error to be returned
func (cmd *Work) DoRun(args []string) (err error) {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	conf.SetLocation(cmd.opts.ConfigFile)

	client, err := NewClient(cmd.ui)
	if err != nil {
		return HandleError(HandleClientError(err, cmd.opts.VerboseOutput), cmd.opts.VerboseOutput)
	}

	var worker *payload.WorkerCreateOutput
	worker, err = client.CreateWorker()
	if err != nil {
		return err
	}

	defer func() {
		err = client.DeleteWorker(worker.WorkerID)
		if err != nil {
			fmt.Printf("failed to delete worker: %v\n", err)
		}
	}()

	fmt.Printf("registered as worker '%s' (project: %s)\n", worker.WorkerID, worker.ProjectID)

	awscreds := naws.NewNerdalizeCredentials(client)
	awssess := session.New(
		aws.NewConfig().WithCredentials(awscreds).WithRegion("eu-west-1"),
	)

	messages := sqs.New(awssess)
	states := sfn.New(awssess)

	//for now, we just parse use the docker cli
	exe, err := exec.LookPath("docker")
	if err != nil {
		return errors.Wrap(err, "failed to find docker executable")
	}

	//receive tasks from the message queue and start the container run loop, it will attemp to create containers for tasks unconditionally if it keeps failing queue retry will backoff. If it succeeds, fails the feedback loop will notify
	go func() {
		for {
			var out *sqs.ReceiveMessageOutput
			if out, err = messages.ReceiveMessage(&sqs.ReceiveMessageInput{
				QueueUrl:        aws.String(worker.QueueURL),
				WaitTimeSeconds: aws.Int64(5),
			}); err != nil {
				fmt.Fprintf(os.Stderr, "failed to receive message: %+v", err)
				//@TODO report async errors
				return
			}

			if len(out.Messages) > 0 {
				for _, msg := range out.Messages {
					task := &payload.Task{}
					if err = json.Unmarshal([]byte(aws.StringValue(msg.Body)), task); err != nil {

						//@TODO throw deserialization errors
						fmt.Fprintf(os.Stderr, "failed to deserialize: %+v", err)
						return
					}

					//@TODO execute a pre-run heartbeat to prevent starting containers for delayed but outdated task tokens. if the heartbeat returns a timed out error don't attempt to start it: (dont forget to delete the message)

					fmt.Fprintf(os.Stderr, "starting task: %s, token: %x\n", task.TaskID, sha1.Sum([]byte(task.ActivityToken)))
					args := []string{
						"run", "-d",
						//@TODO add logging to aws
						fmt.Sprintf("--name=task-%x", sha1.Sum([]byte(task.ActivityToken))),
						fmt.Sprintf("--label=nerd-project=%s", task.ProjectID),
						fmt.Sprintf("--label=nerd-task=%s", task.TaskID),
						fmt.Sprintf("--label=nerd-token=%s", task.ActivityToken),
						fmt.Sprintf("-e=NERD_PROJECT_ID=%s", task.ProjectID),
						fmt.Sprintf("-e=NERD_TASK_ID=%s", task.TaskID),
					}

					if task.InputID != "" {
						args = append(args, fmt.Sprintf("-e=NERD_DATASET_INPUT=%s", task.InputID))
					}

					for key, val := range task.Environment {
						args = append(args, fmt.Sprintf("-e=%s=%s", key, val))
					}

					args = append(args, task.Image)
					cmd := exec.Command(exe, args...)
					cmd.Stderr = os.Stderr
					_ = cmd.Run() //any result is ok

					//delete message, state is persisted in Docker, it is no longer relevant
					if _, err := messages.DeleteMessage(&sqs.DeleteMessageInput{
						QueueUrl:      aws.String(worker.QueueURL),
						ReceiptHandle: msg.ReceiptHandle,
					}); err != nil {
						//@TODO error on return error
						fmt.Fprintf(os.Stderr, "failed to delete message: %+v", err)
						continue
					}
				}
			}
		}
	}()

	//the container loop feeds running task tokens to the feedback loop by polling the `docker ps` output
	pr, pw := io.Pipe()
	psTicker := time.NewTicker(time.Second * 5)
	go func() {
		for range psTicker.C {
			args := []string{"ps", "-a",
				"--no-trunc",
				"--filter=label=nerd-token",
				"--format={{.ID}}\t{{.Status}}\t{{.Label \"nerd-token\"}}\t{{.Label \"nerd-project\"}}\t{{.Label \"nerd-task\"}}",
			}

			cmd := exec.Command(exe, args...)
			cmd.Stdout = pw
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				//@TODO handle errors
				fmt.Fprintln(os.Stderr, "failed to list containers: ", err)
				continue
			}
		}
	}()

	//the scan loop will parse docker states into exit statuses
	scanner := bufio.NewScanner(pr)
	statusCh := make(chan TaskStatus)
	go func() {
		for scanner.Scan() {
			fields := strings.SplitN(scanner.Text(), "\t", 5)
			if len(fields) != 5 {
				statusCh <- TaskStatus{fields[0], fields[2], 255, errors.New("unexpected ps line"), fields[3], fields[4]}
				continue //less then 2 fields, shouldnt happen
			}

			//second field can be interpreted by reversing state .String() https://github.com/docker/docker/blob/b59ee9486fad5fa19f3d0af0eb6c5ce100eae0fc/container/state.go#L70
			status := fields[1]
			if strings.HasPrefix(status, "Up") || strings.HasPrefix(status, "Restarting") || status == "Removal In Progress" || status == "Created" {
				//container is not yet "done": still in progress without statuscode, send heartbeat and continue to next tick
				statusCh <- TaskStatus{fields[0], fields[2], -1, nil, fields[3], fields[4]}
				continue
			} else {
				//container has "exited" or is "dead"
				if status == "Dead" {
					//@See https://github.com/docker/docker/issues/5684
					// There is also a new(ish) container state called "dead", which is set when there were issues removing the container. This is of course a work around for this particular issue, which lets you go and investigate why there is the device or resource busy error (probably a race condition), in which case you can attempt to remove again, or attempt to manually fix (e.g. unmount any left-over mounts, and then remove).
					statusCh <- TaskStatus{fields[0], fields[2], 255, errors.New("failed to remove container"), fields[3], fields[4]}
					continue

				} else if strings.HasPrefix(status, "Exited") {
					right := strings.TrimPrefix(status, "Exited (")
					lefts := strings.SplitN(right, ")", 2)
					if len(lefts) != 2 {
						statusCh <- TaskStatus{fields[0], fields[2], 255, errors.New("unexpected exited format: " + status), fields[3], fields[4]}
						continue
					}

					//write actual status code, can be zero in case of success
					code, err := strconv.Atoi(lefts[0])
					if err != nil {
						statusCh <- TaskStatus{fields[0], fields[2], 255, errors.New("unexpected status code, not a number: " + status), fields[3], fields[4]}
						continue
					} else {
						statusCh <- TaskStatus{fields[0], fields[2], code, nil, fields[3], fields[4]}
						continue
					}

				} else {
					statusCh <- TaskStatus{fields[0], fields[2], 255, errors.New("unexpected status: " + status), fields[3], fields[4]}
					continue
				}
			}
		}
		if err := scanner.Err(); err != nil {
			//@TODO handle scanniong IO errors
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
	}()

	//the feedback loop holds a view of task states and tokens
	for {
		select {
		case <-sigCh: //exit our main loop
			return
		case statusEv := <-statusCh: //sync docker status

			//@TODO start moving logs from the container to cloudwatch

			fmt.Fprintf(os.Stderr, "task-%x is %d\n", sha1.Sum([]byte(statusEv.token)), statusEv.code)

			var err error
			if statusEv.code < 0 {
				fmt.Fprintln(os.Stderr, "heartbeat!")
				_, err = states.SendTaskHeartbeat(&sfn.SendTaskHeartbeatInput{
					TaskToken: aws.String(statusEv.token),
				})
			} else if statusEv.code == 0 {

				var outdata []byte
				if outdata, err = json.Marshal(&payload.TaskResult{
					ProjectID:  statusEv.pid,
					TaskID:     statusEv.tid,
					OutputID:   "d-ffffffff",
					ExitStatus: fmt.Sprintf("Exit Status: %d", statusEv.code),
				}); err != nil {
					fmt.Println("failed to marshal task result: ", err)
					continue
				}

				//success
				fmt.Fprintln(os.Stderr, "success!")
				_, err = states.SendTaskSuccess(&sfn.SendTaskSuccessInput{
					TaskToken: aws.String(statusEv.token),
					Output:    aws.String(string(outdata)),
				})

			} else {
				//failure
				fmt.Fprintln(os.Stderr, "failed!")
				//@TODO dont send cause if .err is nil
				_, err = states.SendTaskFailure(&sfn.SendTaskFailureInput{
					TaskToken: aws.String(statusEv.token),
					Error:     aws.String(fmt.Sprintf(`{"error": "%d"}`, statusEv.code)),
					Cause:     aws.String(fmt.Sprintf(`{"cause": "%v"}`, statusEv.err)),
				})
			}

			if err != nil {
				aerr, ok := err.(awserr.Error)
				if !ok {
					fmt.Println("unexpected non-aws error:", err)
					//@TODO not an aws error, connection issues or otherwise, do not undertake an specific action maybe next time it will be better
					continue
				}

				if aerr.Code() == sfn.ErrCodeTaskTimedOut {
					fmt.Println("aws err:", aerr)
					cmd := exec.Command(exe, "stop", statusEv.cid)
					err = cmd.Run()
					if err != nil {
						fmt.Println("failed to stop task container:", statusEv.cid, statusEv.code, err)
						//@TODO report error
					}

					cmd = exec.Command(exe, "rm", statusEv.cid)
					err = cmd.Run()
					if err != nil {
						fmt.Println("failed to remove timed out task container:", statusEv.cid, statusEv.code, err)
						//@TODO report error
					}
				}
			}
		}
	}
}
