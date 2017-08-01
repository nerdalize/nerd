package command

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/mitchellh/cli"
	v1auth "github.com/nerdalize/nerd/nerd/client/auth/v1"
	"github.com/nerdalize/nerd/nerd/jwt"
	"github.com/nerdalize/nerd/nerd/oauth"
	"github.com/pkg/errors"
)

//WorkloadStartOpts describes command options
type WorkloadStartOpts struct {
	Env          []string `long:"env" short:"e" description:"environment variables"`
	InputDataset string   `long:"input-dataset" short:"d" description:"input dataset ID, will be available in /input in your container"`
	Workers      int      `long:"workers" short:"w" default:"1" description:"number of workers that handle the workload"`
	Instances    int      `long:"instances" short:"i" default:"1" description:"number of working instances"`
	PullSecret   string   `long:"pull-secret" short:"p" description:"the pull secret will be used to fetch the private image"`
}

//WorkloadStart command
type WorkloadStart struct {
	*command
	opts *WorkloadStartOpts
}

//WorkloadStartFactory returns a factory method for the join command
func WorkloadStartFactory() (cli.Command, error) {
	opts := &WorkloadStartOpts{}
	comm, err := newCommand("nerd workload start <image>", "provision a new workload to provide compute", "", opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &WorkloadStart{
		command: comm,
		opts:    opts,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkloadStart) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	//fetching a worker JWT
	authbase, err := url.Parse(cmd.config.Auth.APIEndpoint)
	if err != nil {
		return HandleError(errors.Wrapf(err, "auth endpoint '%v' is not a valid URL", cmd.config.Auth.APIEndpoint))
	}

	authOpsClient := v1auth.NewOpsClient(v1auth.OpsClientConfig{
		Base:   authbase,
		Logger: cmd.outputter.Logger,
	})

	authclient := v1auth.NewClient(v1auth.ClientConfig{
		Base:               authbase,
		Logger:             cmd.outputter.Logger,
		OAuthTokenProvider: oauth.NewConfigProvider(authOpsClient, cmd.config.Auth.ClientID, cmd.session),
	})

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}

	projectID, err := ss.RequireProjectID()
	if err != nil {
		return HandleError(err)
	}

	workerJWT, err := authclient.GetWorkerJWT(projectID, v1auth.NCEScope)
	if err != nil {
		return HandleError(errors.Wrap(err, "failed to get worker JWT"))
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	wenv := map[string]string{}
	for _, l := range cmd.opts.Env {
		split := strings.SplitN(l, "=", 2)
		if len(split) < 2 {
			return HandleError(fmt.Errorf("invalid environment variable format, expected 'FOO=bar' fromat, got: %v", l))
		}
		wenv[split[0]] = split[1]
	}

	wenv[jwt.NerdTokenEnvVar] = workerJWT.Token
	wenv[jwt.NerdSecretEnvVar] = workerJWT.Secret
	configJSON, err := json.Marshal(cmd.config)
	if err != nil {
		return HandleError(errors.Wrap(err, "failed to marshal config"))
	}
	wenv[EnvConfigJSON] = string(configJSON)
	wenv[EnvNerdProject] = ss.Project.Name

	workload, err := bclient.CreateWorkload(ss.Project.Name, args[0], cmd.opts.InputDataset, cmd.opts.PullSecret, wenv, cmd.opts.Instances, true)
	if err != nil {
		return HandleError(err)
	}

	cmd.outputter.Logger.Printf("Workload created with ID: %s", workload.WorkloadID)
	return nil
}
