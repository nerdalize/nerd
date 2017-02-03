package command

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd"
)

//WorkOpts describes command options
type WorkOpts struct {
	AWSQueueURL string `long:"aws-queue-url" required:"true" description:"url of the aws sqs queue"`
	*NerdAPIOpts
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

//DoRun is called by run and allows an error to be returned
func (cmd *Work) DoRun(args []string) (err error) {

	//setup aws session
	//@TODO use boris aws credentials user fetching
	sess, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("failed to setup AWS session: %v", err)
	}

	//use sqs as our queue backend
	mq := sqs.New(sess)
	qurl := cmd.opts.AWSQueueURL
	wtime := int64(5)

	//star long polling sqs queue
	cmd.command.ui.Info(fmt.Sprintf("long-polling '%s'", qurl))
	for {
		msgs, err := mq.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(qurl),
			MaxNumberOfMessages: aws.Int64(1),
			WaitTimeSeconds:     aws.Int64(wtime),
		})
		if err != nil {
			return fmt.Errorf("failed to poll for tasks: %v", err)
		}

		if len(msgs.Messages) > 0 {
			go handleTask(qurl, mq, msgs.Messages[0], wtime, cmd.opts)
		}
	}
}

func patchTaskStatus(taskID string, ts *nerd.TaskStatus, opts *WorkOpts) (err error) {
	loc, err := opts.URL("/tasks/" + taskID)
	if err != nil {
		return fmt.Errorf("failed to create API url from cli options: %+v", err)
	}

	log.Printf("patching task to %s", loc)
	body := bytes.NewBuffer(nil)
	enc := json.NewEncoder(body)
	err = enc.Encode(ts)
	if err != nil {
		return fmt.Errorf("failed to encode provided task status: %v", err)
	}

	req, err := http.NewRequest("PATCH", loc.String(), body)
	if err != nil {
		return fmt.Errorf("failed to create API request: %+v", err)
	}

	//@TODO abstract into a default http client
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("API request '%s %s' failed: %v", req.Method, loc, err)
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API request '%s %s' returned unexpected status from API: %v", req.Method, loc, resp.Status)
	}

	//@TODO find a more user friendly way of returning info from the API
	_, err = io.Copy(os.Stderr, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to output API response: %v", err)
	}

	log.Printf("patched task '%s' with a new status", taskID)

	return nil
}

func handleTask(qurl string, mq *sqs.SQS, msg *sqs.Message, timeout int64, opts *WorkOpts) {
	log.Printf("received task '%v' now in-flight, invisible for: %vs)", aws.StringValue(msg.Body), timeout)

	//for now simply start a docker run command for each task
	t := &nerd.Task{}
	err := json.Unmarshal([]byte(aws.StringValue(msg.Body)), t)
	if err != nil {
		log.Printf("failed to decode task: %v", err)
		return
	}

	loglines := []string{}
	pr, pw := io.Pipe()
	go func() {
		scanner := bufio.NewScanner(pr)
		for scanner.Scan() {
			loglines = append(loglines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
		}
	}()

	//@TODO without filtering for privileged, this is obviously very dangerous
	t.Args = append([]string{"run"}, t.Args...)
	t.Args = append(t.Args, t.Image)
	cmd := exec.Command("docker", t.Args...)

	cmd.Stdout = pw
	cmd.Stderr = pw
	err = cmd.Start()
	if err != nil {
		log.Printf("failed to start docker container: %v", err)
		return
	}

	go func() {
		f := time.Duration(timeout) * time.Second
		for range time.Tick(f) {

			//create new status
			taskStatus := &nerd.TaskStatus{
				Status: "running",
				Logs:   loglines,
			}

			if cmd.ProcessState != nil {
				taskStatus.Status = cmd.ProcessState.String()
			}

			//always update the task
			log.Printf("task status is: '%v', patching...", taskStatus)
			err = patchTaskStatus(aws.StringValue(msg.MessageId), taskStatus, opts)
			if err != nil {
				fmt.Printf("Failed to patch task with status: %v", err)
			}

			//as long as the process state is nil, it has not finished and we want to prologin the invisibility
			if cmd.ProcessState != nil {
				break
			}

			log.Printf(
				"%v still executing, updating visibility +%vs",
				time.Now(),
				f,
			)

			//update visibility
			if _, err := mq.ChangeMessageVisibility(&sqs.ChangeMessageVisibilityInput{
				QueueUrl:          aws.String(qurl),
				ReceiptHandle:     msg.ReceiptHandle,
				VisibilityTimeout: aws.Int64(timeout),
			}); err != nil {
				log.Printf("failed to change visibility: %v", err)
				break
			}

			log.Printf("tick process %v %v", cmd.Process, cmd.ProcessState)
		}
	}()

	err = cmd.Wait()
	if err != nil {
		log.Printf("docker container failed: %v", err)
		return
	}

	log.Printf("done executing task, deleting message...")
	if _, err := mq.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(qurl),
		ReceiptHandle: msg.ReceiptHandle,
	}); err != nil {
		log.Printf("failed to delete message: %v", err)
	}

	err = pw.Close()
	if err != nil {
		log.Printf("error closing scanner: %v", err)
	}
}
