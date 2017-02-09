package command

import (
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd"
	"github.com/nerdalize/nerd/nerd/client"
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
func (cmd *Run) DoRun(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	user := nerd.GetCurrentUser()
	var akey string
	var skey string
	if user != nil {
		creds, err := user.GetAWSCredentials()
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

	c := client.NewNerdAPI(cmd.opts.NerdAPIConfig())

	err := c.CreateTask(args[0], args[1], akey, skey, args[2:])
	if err != nil {
		return fmt.Errorf("failed to post to nerdalize API: %v", err)
	}

	return nil
}
