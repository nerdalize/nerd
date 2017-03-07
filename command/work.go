package command

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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
		return err
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

MAINLOOP:
	for {
		select {
		case task := <-taskCh:
			//@TODO docker interaction logic goes here
			fmt.Fprintf(os.Stderr, "task: %+v", task)
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
