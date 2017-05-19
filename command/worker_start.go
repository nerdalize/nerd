package command

import (
	"fmt"
	"net/url"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/nerd/oauth"
	"github.com/pkg/errors"
)

//WorkerStartOpts describes command options
type WorkerStartOpts struct {
	NerdOpts
}

//WorkerStart command
type WorkerStart struct {
	*command
	opts   *WorkerStartOpts
	parser *flags.Parser
}

//WorkerStartFactory returns a factory method for the join command
func WorkerStartFactory() (cli.Command, error) {
	cmd := &WorkerStart{
		command: &command{
			help:     "",
			synopsis: "provision a new worker to provide compute",
			parser:   flags.NewNamedParser("nerd worker start", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &WorkerStartOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkerStart) DoRun(args []string) (err error) {
	// return fmt.Errorf("not yet implemented")
	config, err := conf.Read()
	if err != nil {
		HandleError(err)
	}
	authbase, err := url.Parse(config.Auth.APIEndpoint)
	if err != nil {
		HandleError(errors.Wrapf(err, "auth endpoint '%v' is not a valid URL", config.Auth.APIEndpoint))
	}
	authOpsClient := v1auth.NewOpsClient(v1auth.OpsClientConfig{
		Base:   authbase,
		Logger: logrus.StandardLogger(),
	})
	authclient := v1auth.NewClient(v1auth.ClientConfig{
		Base:               authbase,
		Logger:             logrus.StandardLogger(),
		OAuthTokenProvider: oauth.NewConfigProvider(authOpsClient),
	})
	batchclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err)
	}

	workerJWT, err := authclient.GetWorkerJWT(config.CurrentProject.Name, v1auth.NCEScope)
	if err != nil {
		HandleError(errors.Wrap(err, "failed to get worker JWT"))
	}
	input := &v1payload.CreateWorkerInput{
		ProjectID: config.CurrentProject.Name,
		Env: map[string]string{
			"NERD_JWT":        workerJWT.Token,
			"NERD_JWT_SECRET": workerJWT.Secret,
		},
	}
	output, err := batchclient.CreateWorker(config.CurrentProject.Name, input)
	if err != nil {
		HandleError(errors.Wrap(err, "failed to start worker"))
	}
	fmt.Println(output)
	return nil
}
