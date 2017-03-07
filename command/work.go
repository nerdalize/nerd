package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
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

//DoRun is called by run and allows an error to be returned
func (cmd *Work) DoRun(args []string) (err error) {
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	home, err := homedir.Dir()
	if err != nil {
		return errors.Wrap(err, "failed to determin home directory")
	}

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

	fmt.Fprintf(os.Stderr, "identified as worker '%s'", worker.WorkerID)

	//@TODO create worker if no local state is available?

	awscreds := data.NewNerdalizeCredentials(api)
	awssess := session.New(
		aws.NewConfig().WithCredentials(awscreds).WithRegion("eu-west-1"),
	)

	//the taskch receives tasks from the message queue
	taskCh := make(chan *payload.Task)
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

						//@TODO return deserialization errors
						fmt.Fprintf(os.Stderr, "failed to deserialize: %+v", err)
						return
					}

					taskCh <- task
				}
			}
		}
	}()

	states := sfn.New(awssess)

	tasks := map[string]struct{}{}
MAINLOOP:
	for {
		select {
		case task := <-taskCh:
			if _, ok := tasks[task.TaskID]; ok {
				continue //already managed, skip
			}

			//@TODO we're keeping state here, is that OK
			tasks[task.TaskID] = struct{}{}

			//the task routine
			go func() {
				ticker := time.Tick(time.Second * 5)
				for range ticker {

					exe, err := exec.LookPath("docker")
					if err != nil {
						//@TODO mark as failure, unlikely the exe will ever be resolved
						break
					}

					name := fmt.Sprintf("%s_%s", task.ProjectID, task.TaskID)
					args := []string{
						"run",
						"-d",
						fmt.Sprintf("--name=%s", name),
						fmt.Sprintf("--label=task-token=%s", task.ActivityToken),
						task.Image,
					}

					//@TODO what do we do with stdout?
					runcmd := exec.Command(exe, args...)
					runcmd.Stderr = os.Stderr
					runcmd.Stdout = os.Stderr
					_ = runcmd.Start()
					//@TODO we dont actually care if this succeeds

					// if err != nil {
					// 	fmt.Fprintf(os.Stderr, "couldnt start: %+v", err)
					// 	//@TODO determine if:
					// 	//	- still running: heartbeat
					// 	//  - not running:
					// 	//   - if failed: send failure
					// 	//   - if success: send success
					// }

					buf := bytes.NewBuffer(nil)
					args = []string{"inspect", name}
					inscmd := exec.Command(exe, args...)
					// inscmd.Stderr = os.Stderr
					inscmd.Stdout = buf
					_ = inscmd.Run()

					//@SEE https://github.com/docker/docker/blob/b59ee9486fad5fa19f3d0af0eb6c5ce100eae0fc/container/state.go#L17
					status := []struct {
						State struct {
							Status     string
							Running    bool
							Paused     bool
							Restarting bool
							OOMKilled  bool
							Dead       bool
							Pid        int
							ExitCode   int
							Error      string
							StartedAt  time.Time
							FinishedAt time.Time
						}
					}{}

					err = json.Unmarshal(buf.Bytes(), &status)
					if err != nil || len(status) < 1 {
						fmt.Fprintf(os.Stderr, "failed to unmarshal: %+v\n", buf.String())
						continue
					}

					//@TODO if task timed out, cancel container

					if status[0].State.Running {
						fmt.Fprintf(os.Stderr, "task '%s' is still running, sending hb\n", name)
						_, err = states.SendTaskHeartbeat(&sfn.SendTaskHeartbeatInput{
							TaskToken: aws.String(task.ActivityToken),
						})
						if err != nil {
							aerr, ok := err.(awserr.Error)
							if !ok || aerr.Code() != sfn.ErrCodeTaskTimedOut {
								fmt.Fprintf(os.Stderr, "failed to send hb for %s: %+v\n", name, err)
								continue
							}

							//task timed out, nothing to do
							break
						}
						//heartbeat
					} else {
						fmt.Fprintf(os.Stderr, "task '%s' is NOT running: %+v\n", name, status[0].State)

						if status[0].State.ExitCode != 0 {
							if _, err = states.SendTaskFailure(&sfn.SendTaskFailureInput{
								Cause:     aws.String(status[0].State.Error),
								Error:     aws.String(fmt.Sprintf("exit code: %d", status[0].State.ExitCode)),
								TaskToken: aws.String(task.ActivityToken),
							}); err != nil {
								aerr, ok := err.(awserr.Error)
								if !ok || aerr.Code() != sfn.ErrCodeTaskTimedOut {
									fmt.Fprintf(os.Stderr, "failed to send failure for %s: %+v\n", name, err)
									continue
								}

								//task timed out, nothing to do
								break
							}
						} else {
							if _, err = states.SendTaskSuccess(&sfn.SendTaskSuccessInput{
								TaskToken: aws.String(task.ActivityToken),
								Output:    aws.String(`{"foo": "bar"}`), //@TODO send task with output
							}); err != nil {
								aerr, ok := err.(awserr.Error)
								if !ok || aerr.Code() != sfn.ErrCodeTaskTimedOut {
									fmt.Fprintf(os.Stderr, "failed to send success for %s: %+v\n", name, err)
									continue
								}

								//task timed out, nothing to do
								break
							}
						}

						break
					}
				}
			}()

		case <-sigCh:
			break MAINLOOP
		}
	}

	err = api.DeleteWorker(worker.WorkerID)
	if err != nil {
		return err
	}

	return nil
}
