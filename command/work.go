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
type WorkOpts struct{}

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
			parser:   flags.NewNamedParser("nerd upload", flags.Default),
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
	qurl := "https://sqs.eu-west-1.amazonaws.com/469576599204/remove-me"

	//star long polling sqs queue
	for {
		msgs, err := mq.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(qurl),
			MaxNumberOfMessages: aws.Int64(1), //receive only one message at a time
			WaitTimeSeconds:     aws.Int64(5), //after which the q considers failed
		})
		if err != nil {
			return fmt.Errorf("failed to poll for tasks: %v", err)
		}

		if len(msgs.Messages) > 0 {
			msg := msgs.Messages[0]
			//We allow ourselves to perform work here, at this time the task is invisible to other receivers. If the messages is deleted using the receipt handle the message is consider "handled". If the the WaitTimeSeconds elapses the queue will consider the receive a failure and makes the message available to other receivers.
			log.Printf("received msg: %v", msg)
			if strings.Contains(aws.StringValue(msg.Body), "nerd") {
				//if the message contains "nerd" we cannot handle this message and change the visibility back to "visible"; maybe other receivers can handle it.

				cout, err := mq.ChangeMessageVisibility(&sqs.ChangeMessageVisibilityInput{
					QueueUrl:          aws.String(qurl),
					ReceiptHandle:     msg.ReceiptHandle,
					VisibilityTimeout: aws.Int64(0), //new visibility to: 0 = release
				})
				if err != nil {
					return fmt.Errorf("failed to delete message: %v", err)
				}

				log.Printf("change message visibility: %v", cout)
				continue
			}

			time.Sleep(3 * time.Second)
			log.Printf("done handling, deleting task...")
			dout, err := mq.DeleteMessage(&sqs.DeleteMessageInput{
				QueueUrl:      aws.String(qurl),
				ReceiptHandle: msg.ReceiptHandle,
			})

			if err != nil {
				return fmt.Errorf("failed to delete message: %v", err)
			}

			log.Printf("deleted message: %v", dout)
		}

	}
}
