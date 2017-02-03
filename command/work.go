package command

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//WorkOpts describes command options
type WorkOpts struct {
	AWSQueueURL string `long:"aws-queue-url" required:"true" description:"url of the aws sqs queue"`
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
			go handleTask(qurl, mq, msgs.Messages[0], wtime)
		}
	}
}

func handleTask(qurl string, mq *sqs.SQS, msg *sqs.Message, timeout int64) {
	log.Printf("received task '%v' now in-flight, invisible for: %vs)", aws.StringValue(msg.Body), timeout)

	//to illustrate, if a message contains "nerd" we will handle it and succeed if it doesnt we do nothing "letting it timeout"
	if strings.Contains(aws.StringValue(msg.Body), "nerd") {

		log.Printf("task contains nerd! lets handle it")
		ticker := time.Tick(3 * time.Second)
		tickN := 0
		for range ticker {
			tickN++
			if tickN > 5 {
				break
			}

			log.Printf(
				"%v still executing (%v), updating visibility +%vs",
				time.Now(),
				tickN,
				timeout,
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
		}

		//deleting message
		log.Printf("Done executing task, deleting message")
		if _, err := mq.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      aws.String(qurl),
			ReceiptHandle: msg.ReceiptHandle,
		}); err != nil {
			log.Printf("failed to delete message: %v", err)
		}
	}
}
