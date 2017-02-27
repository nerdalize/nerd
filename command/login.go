package command

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
)

//LoginOpts describes command options
type LoginOpts struct {
	*NerdAPIOpts
	*OutputOpts
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
	// if len(args) < 1 {
	// 	return fmt.Errorf("not enough arguments, see --help")
	// }
	//
	// c := client.NewNerdAPI(cmd.opts.NerdAPIConfig())
	//
	// sess, err := c.CreateSession(args[0])
	// if err != nil {
	// 	return HandleError(HandleClientError(err, cmd.opts.VerboseOutput), cmd.opts.VerboseOutput)
	// }
	//
	// fmt.Println("AWS_ACCESS_KEY_ID=" + sess.AWSAccessKeyID)
	// fmt.Println("AWS_SECRET_ACCESS_KEY=" + sess.AWSSecretAccessKey)
	// fmt.Println("AWS_SESSION_TOKEN=" + sess.AWSSessionToken)
	// fmt.Println("AWS_STORAGE_BUCKET=" + sess.AWSStorageBucket)
	// fmt.Println("AWS_STORAGE_ROOT=" + sess.AWSStorageRoot)

	return nil
}
