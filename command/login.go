package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd"
)

//LoginOpts describes command options
type LoginOpts struct {
	*NerdAPIOpts
}

//Login command
type Login struct {
	*command

	ui     cli.Ui
	opts   *LoginOpts
	parser *flags.Parser
}

//LoginFactory returns a factory method for the join command
func LoginFactory() func() (cmd cli.Command, err error) {
	cmd := &Login{
		command: &command{
			help:     "",
			synopsis: "setup an authorized session for the cloud",
			parser:   flags.NewNamedParser("nerd login", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &LoginOpts{},
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
func (cmd *Login) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	loc, err := cmd.opts.URL("/sessions/" + args[0])
	if err != nil {
		return fmt.Errorf("failed to create API url from cli options: %+v", err)
	}

	req, err := http.NewRequest("POST", loc.String(), bytes.NewBufferString(`{}`))
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

	sess := &nerd.Session{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(sess)
	if err != nil {
		return fmt.Errorf("failed to deserialize session: %v", err)
	}

	fmt.Println("AWS_ACCESS_KEY_ID=" + sess.AWSAccessKeyID)
	fmt.Println("AWS_SECRET_ACCESS_KEY=" + sess.AWSSecretAccessKey)
	fmt.Println("AWS_SQS_QUEUE_URL=" + sess.AWSSQSQueueURL)
	fmt.Println("AWS_REGION=" + sess.AWSRegion)

	return nil
}
