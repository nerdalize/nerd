package command

import (
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
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
	homedir "github.com/mitchellh/go-homedir"
	"github.com/nerdalize/nerd/nerd/client"
	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/nerdalize/nerd/nerd/data"
	"github.com/nerdalize/nerd/nerd/payload"
	"github.com/pkg/errors"
)

//WorkOpts describes command options
type WorkOpts struct {
	*OutputOpts
}

//Work command
type Work struct {
	*command

	ui     cli.Ui
	opts   *WorkOpts
	parser *flags.Parser
}

//WorkFactory returns a factory method for the join command
func WorkFactory() func() (cmd cli.Command, err error) {
	cmd := &Work{
		command: &command{
			help:     "",
			synopsis: "handle tasks using local compute resources",
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

type staticToken string

func (t staticToken) IsExpired() bool {
	return true
}

func (t staticToken) Retrieve() (*credentials.NerdAPIValue, error) {
	return &credentials.NerdAPIValue{
		NerdToken: string(t),
	}, nil
}

//TaskPK uniquely identifies a task on a node
type TaskPK struct {
	pid string //project id
	tid string //task id
}

func (pk TaskPK) String() string {
	return fmt.Sprintf("%s_%s", pk.pid, pk.tid)
}

//TaskToken allow feedback about a task
type TaskToken struct {
	TaskPK
	token string //task activity token
}

//TaskStatus describes the container status
type TaskStatus struct {
	cid string //container id
	TaskPK
	code int   //exit code
	err  error //application error
}

//Task holds all local information of a task
type Task struct {
	TaskPK
	*TaskToken
	*TaskStatus
	timeoutN int
}

func (t *Task) String() string {
	return fmt.Sprintf("%s (%d)", t.TaskPK, t.TaskStatus.code)
}

//DoRun is called by run and allows an error to be returned
func (cmd *Work) DoRun(args []string) (err error) {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	home, err := homedir.Dir()
	if err != nil {
		return errors.Wrap(err, "failed to determin home directory")
	}

	// t-096054f3 AAAAKgAAAAIAAAAAAAAAATVEhfvQqWZIB1VJWhfj+CPkQeqJrEGOLrm1BD8tCDxr+UyNuQAFG7mmtfmy4hG12cJprYiXNEq8j3dyfugvOrsjma42GwbzW+41YwJcmMIb47F/cvvpT+qNQgHbBjP3Mc3vIWpPTe8AWC/z5WugqgPyx5ekbbYAow5qPE4haXkmnuADJRB6exwg0qQa8IkryZ7zYWfCPiQZI2WwreNOmhEJscRqmBbNZAn9waBs+aBlcmq/BEz5qXxwVpfCS0whFPg5u42gOwWImcbe3O5zSaIByq4sox6wCPE8lFuDm7P6efqOQcC5jhY0Lq2EbOilIPRDh1XIfwLQh4WOKI8YME7RxG9hxGt8mXKXZFd5+KqCWhKqQtdrzFR5+QCHL6Szsh+7k2CcMJ2C0ySWq6xSDZjW9rC3IvUWtd47w8NbLcpbr2QOZGekqztzCKHT5b2DqlvWLo+hcxqLkugSoixhP4QliYYfMTAbtzkAjUx9cXXXUkrUzhIAf5HNqptLTrfSknV8oXcQmhtJtd+Kk4yETADBJIZXqo+DuIG0g469Hikc

	// t-1e619f00 AAAAKgAAAAIAAAAAAAAAAcY3enG5WaFwpCZUDiKi8Wv+730Kdj0hJPQMRagD5CR76kuNUoriAtExsw3yUmvgiZNdk/E0xeeZ/uCxcQjiR+sAV3tWHmphZWYurcs2opHPyyUXHLzSVa8mlpYjER3DVzN3kfgMfrX7rf8aFtBM7XEdu2phFKg815O//Fq4niOXUkHq1UJhqhu9EyfY2lKEFgg/Wclm3atjXhcOzilHqvmIqFtrhKGmuM9hxSfq+/N5ecJLEHhQBzk8Qz5GGC0mFOjhlsqM/RKt1H3/2wJR0I/yhL0KnIOtaitPD18S1/AVP9yWwMIC2JB98tnO8mJcSVM8dOx3nNCGHKC71B1p+/z5l+sKfWuwfc6pR76q90/3LAN878lRCG/KJFzZOAMSaudd+ckpU0xHoIXtY7GFIIicdHerN2mURet76WwXuW+z/XvNWko0AExFY3AURG4xCOcPQHrc9lOH9UDJVoAFLnMXRKUJ9OOV5qQ8/c0Tu/Nz4CL1vvkobEBR+x0bvbhoS/xqRAnTiZjmkhWyqGBx44H827ax1tkhe8RJrOgEQ+xY

	jwtp := filepath.Join(home, ".nerd", "token")
	jwtd, err := ioutil.ReadFile(jwtp)
	if err != nil {
		return errors.Wrap(err, "failed to read json web token file")
	}

	//@TODO instead of static token, read from config file
	p := staticToken(string(jwtd))
	creds := credentials.NewNerdAPI(p)
	api, err := client.NewNerdAPI(creds)
	if err != nil {
		return fmt.Errorf("%+v", err)
	}

	//@TODO use existing worker id if configured
	var worker *payload.WorkerCreateOutput
	if true {
		worker, err = api.CreateWorker()
		if err != nil {
			return err
		}
	}

	defer func() {
		err = api.DeleteWorker(worker.WorkerID)
		if err != nil {
			//@TODO report worker delete error
			// return err
		}
	}()

	fmt.Fprintf(os.Stderr, "identified as worker '%s'\n", worker.WorkerID)

	//@TODO create worker if no local state is available?

	awscreds := data.NewNerdalizeCredentials(api)
	awssess := session.New(
		aws.NewConfig().WithCredentials(awscreds).WithRegion("eu-west-1"),
	)

	//
	// The logic below should be merged into master
	//

	// tasks := map[string]

	//for now, we just parse use the docker cli
	exe, err := exec.LookPath("docker")
	if err != nil {
		return errors.Wrap(err, "failed to find docker executable")
	}

	//receive tasks from the message queue and start the container run loop, it will attemp to create containers for tasks unconditionally if it keeps failing queue retry will backoff. If it succeeds, fails the feedback loop will notify
	tokenCh := make(chan TaskToken)
	go func() {
		messages := sqs.New(awssess)
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

					// fmt.Fprintf(os.Stderr, "message: %s %x \n", task.TaskID, sha1.Sum([]byte(task.ActivityToken)))
					//a new token has arrived, update our local view of the state
					tokenCh <- TaskToken{TaskPK{task.ProjectID, task.TaskID}, task.ActivityToken}

					name := fmt.Sprintf("%s_%s", task.ProjectID, task.TaskID)
					args := []string{
						"run", "-d",
						//@TODO add logging to aws
						fmt.Sprintf("--name=%s", name),
						fmt.Sprintf("--label=nerd-project=%s", task.ProjectID),
						fmt.Sprintf("--label=nerd-task=%s", task.TaskID),
						task.Image,
					}

					cmd := exec.Command(exe, args...)
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
				"--filter=label=nerd-task",
				"--format={{.ID}}\t{{.Status}}\t{{.Label \"nerd-project\"}}\t{{.Label \"nerd-task\"}}",
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
			fields := strings.SplitN(scanner.Text(), "\t", 4)
			if len(fields) != 4 {
				statusCh <- TaskStatus{fields[0], TaskPK{fields[2], fields[3]}, 255, errors.New("unexpected ps line")}
				continue //less then 2 fields, shouldnt happen
			}

			//second field can be interpreted by reversing state .String() https://github.com/docker/docker/blob/b59ee9486fad5fa19f3d0af0eb6c5ce100eae0fc/container/state.go#L70
			status := fields[1]
			if strings.HasPrefix(status, "Up") || strings.HasPrefix(status, "Restarting") || status == "Removal In Progress" || status == "Created" {
				//container is not yet "done": still in progress without statuscode, send heartbeat and continue to next tick
				statusCh <- TaskStatus{fields[0], TaskPK{fields[2], fields[3]}, -1, nil}
				continue
			} else {
				//container has "exited" or is "dead"
				if status == "Dead" {
					//@See https://github.com/docker/docker/issues/5684
					// There is also a new(ish) container state called "dead", which is set when there were issues removing the container. This is of course a work around for this particular issue, which lets you go and investigate why there is the device or resource busy error (probably a race condition), in which case you can attempt to remove again, or attempt to manually fix (e.g. unmount any left-over mounts, and then remove).
					statusCh <- TaskStatus{fields[0], TaskPK{fields[2], fields[3]}, 255, errors.New("failed to remove container")}
					continue

				} else if strings.HasPrefix(status, "Exited") {
					right := strings.TrimPrefix(status, "Exited (")
					lefts := strings.SplitN(right, ")", 2)
					if len(lefts) != 2 {
						statusCh <- TaskStatus{fields[0], TaskPK{fields[2], fields[3]}, 255, errors.New("unexpected exited format: " + status)}
						continue
					}

					//write actual status code, can be zero in case of success
					code, err := strconv.Atoi(lefts[0])
					if err != nil {
						statusCh <- TaskStatus{fields[0], TaskPK{fields[2], fields[3]}, 255, errors.New("unexpected status code, not a number: " + status)}
						continue
					} else {
						statusCh <- TaskStatus{fields[0], TaskPK{fields[2], fields[3]}, code, nil}
						continue
					}

				} else {
					statusCh <- TaskStatus{fields[0], TaskPK{fields[2], fields[3]}, 255, errors.New("unexpected status: " + status)}
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
	states := sfn.New(awssess)
	tasks := map[string]*Task{}
	send := func() {
		for _, task := range tasks {
			if task.TaskStatus == nil {
				continue
			}

			if task.timeoutN > 2 {

				//@TODO task timeout could mean the task was cancelled or that a new activity token was sendout in the meantime
				fmt.Fprintf(os.Stderr, "task '%+v' timed out, removing container...\n", task.TaskPK)

				cmd := exec.Command(exe, "stop", task.cid)
				err = cmd.Run()
				if err != nil {
					fmt.Println("failed to stop task container:", task.cid, task.code, err)
					//@TODO report error
				}

				cmd = exec.Command(exe, "rm", task.cid)
				err = cmd.Run()
				if err != nil {
					fmt.Println("failed to remove timed out task container:", task.cid, task.code, err)
					//@TODO report error
				}

				delete(tasks, task.TaskPK.String())
				continue
			}

			if task.TaskToken == nil {
				task.timeoutN = task.timeoutN + 1
				continue
			}

			var err error
			if task.code < 0 {
				fmt.Fprintln(os.Stderr, "heartbeat:", task.TaskPK)
				_, err = states.SendTaskHeartbeat(&sfn.SendTaskHeartbeatInput{
					TaskToken: aws.String(task.token),
				})
			} else if task.code == 0 {
				//success
				fmt.Fprintln(os.Stderr, "success:", task.TaskPK)
				_, err = states.SendTaskSuccess(&sfn.SendTaskSuccessInput{
					TaskToken: aws.String(task.token),
					Output:    aws.String(`{"foo": "bar"}`),
				})

			} else {
				//failure
				fmt.Fprintln(os.Stderr, "failed:", task.TaskPK)
				_, err = states.SendTaskFailure(&sfn.SendTaskFailureInput{
					TaskToken: aws.String(task.token),
					Error:     aws.String(`{"error": "foo"}`),
					Cause:     aws.String(`{"cause": "bar"}`),
				})
			}

			aerr, ok := err.(awserr.Error)
			if !ok {
				//@TODO not an aws error, connection issues or otherwise
				continue
			}

			//the activity token is no longer valid, this could be that a newer activity token has replaced it, schedule for destruction
			if aerr.Code() == sfn.ErrCodeTaskTimedOut {
				fmt.Fprintf(os.Stderr, "task %s timed out, token hash: %x\n", task.TaskPK, sha1.Sum([]byte(task.token)))
				task.timeoutN = task.timeoutN + 1
			} else {
				task.timeoutN = 0
			}
		}
	}

	//The main loop consolidates token events and status events into our task mapping and triggers the send function in order to update the remote. It needs to run quickly enough such that it send heartbeats with new tokens before they expire (timeout)
	sendTicker := time.NewTicker(time.Second * 10)
MAINLOOP:
	for {
		select {
		case <-sigCh: //exit our main loop
			break MAINLOOP
		case statusEv := <-statusCh: //sync docker status
			existing, ok := tasks[statusEv.TaskPK.String()]
			if ok && existing != nil {
				existing.TaskStatus = &statusEv
			} else {
				existing = &Task{
					TaskPK:     statusEv.TaskPK,
					TaskStatus: &statusEv,
				}
			}

			tasks[statusEv.TaskPK.String()] = existing
			fmt.Fprintf(os.Stderr, "new state for %+v code: %d\n", statusEv.TaskPK, statusEv.code)
		case tokenEv := <-tokenCh: //sync with new tokens
			existing, ok := tasks[tokenEv.TaskPK.String()]
			if ok && existing != nil {
				existing.TaskToken = &tokenEv
			} else {
				existing = &Task{
					TaskPK:    tokenEv.TaskPK,
					TaskToken: &tokenEv,
				}
			}

			fmt.Fprintf(os.Stderr, "new token for %+v tokenHash: %x\n", tokenEv.TaskPK, sha1.Sum([]byte(tokenEv.token)))
		case <-sendTicker.C: //sync remote
			send()
		}
	}

	// 		var err error
	// 		if status.code < 0 {
	// 			fmt.Fprintln(os.Stderr, "heartbeat:", status.TaskPK)
	// 			_, err = states.SendTaskHeartbeat(&sfn.SendTaskHeartbeatInput{
	// 				TaskToken: aws.String(status.token),
	// 			})
	// 		} else if status.code == 0 {
	// 			//success
	// 			fmt.Fprintln(os.Stderr, "success:", status.TaskPK)
	// 			_, err = states.SendTaskSuccess(&sfn.SendTaskSuccessInput{
	// 				TaskToken: aws.String(status.token),
	// 				Output:    aws.String(`{"foo": "bar"}`),
	// 			})
	//
	// 		} else {
	// 			//failure
	// 			fmt.Fprintln(os.Stderr, "failed:", status.TaskPK)
	// 			_, err = states.SendTaskFailure(&sfn.SendTaskFailureInput{
	// 				TaskToken: aws.String(status.token),
	// 				Error:     aws.String(`{"error": "foo"}`),
	// 				Cause:     aws.String(`{"cause": "bar"}`),
	// 			})
	// 		}
	//
	// 		aerr, ok := err.(awserr.Error)
	// 		if !ok {
	// 			//@TODO not an aws error, connection issues or otherwise
	// 			continue
	// 		}
	//
	// 		//the activity token is no longer valid, this could be that a newer activity token has replaced it, schedule for destruction
	// 		if aerr.Code() == sfn.ErrCodeTaskTimedOut {
	//
	// 			//@TODO task timeout could mean the task was cancelled or that a new activity token was sendout in the meantime
	// 			fmt.Fprintf(os.Stderr, "task '%+v' timed out, removing container...\n", status.TaskPK)
	//
	// 			cmd := exec.Command(exe, "stop", status.cid)
	// 			err = cmd.Run()
	// 			if err != nil {
	// 				fmt.Println("failed to stop task container:", status.cid, status.code, err)
	// 				//@TODO report error
	// 			}
	//
	// 			cmd = exec.Command(exe, "rm", status.cid)
	// 			err = cmd.Run()
	// 			if err != nil {
	// 				fmt.Println("failed to remove timed out task container:", status.cid, status.code, err)
	// 				//@TODO report error
	// 			}
	// 		}
	// 	}
	// }()

	return nil
}
