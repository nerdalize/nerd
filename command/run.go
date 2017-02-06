package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd"
)

//RunOpts describes command options
type RunOpts struct {
	*NerdAPIOpts
}

//Run command
type Run struct {
	*command

	ui     cli.Ui
	opts   *RunOpts
	parser *flags.Parser
}

//RunFactory returns a factory method for the join command
func RunFactory() func() (cmd cli.Command, err error) {
	cmd := &Run{
		command: &command{
			help:     "",
			synopsis: "create a new compute task for a dataset",
			parser:   flags.NewNamedParser("nerd run <image> <dataset>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &RunOpts{},
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
func (cmd *Run) DoRun(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	loc, err := cmd.opts.URL("/tasks")
	if err != nil {
		return fmt.Errorf("failed to create API url from cli options: %+v", err)
	}

	user := nerd.GetCurrentUser()
	var akey string
	var skey string
	var creds *credentials.Credentials
	if user != nil {
		creds, err = user.GetAWSCredentials()
		if err != nil {
			return fmt.Errorf("failed to get user credentials: %v", err)
		}

		keys, err := creds.Get()
		if err != nil {
			return fmt.Errorf("failed to get access key from credentials: %v", err)
		}

		akey = keys.AccessKeyID
		skey = keys.SecretAccessKey
	}

	args = append(args, "-e=DATASET="+args[1])
	args = append(args, "-e=AWS_ACCESS_KEY_ID="+akey)
	args = append(args, "-e=AWS_SECRET_ACCESS_KEY="+skey)

	log.Printf("submitting task to %s", loc)
	body := bytes.NewBuffer(nil)
	enc := json.NewEncoder(body)
	err = enc.Encode(&nerd.Task{
		Image:   args[0],
		Dataset: args[1],
		Args:    args[2:],
	})
	if err != nil {
		return fmt.Errorf("failed to encode provided task definition: %v", err)
	}

	req, err := http.NewRequest("POST", loc.String(), body)
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

	return nil
}
