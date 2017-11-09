package command

import (
	"context"
	"net/http"
	"time"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

	clusterSvc "github.com/nerdalize/nerd/nerd/client/cluster/svc"
)

//JobCreateOpts describes command options
type JobCreateOpts struct {
	Name string `long:"name" default:"" default-mask:"" description:"Provide a unique name for this job"`
}

//JobCreate command
type JobCreate struct {
	*command
	opts *JobCreateOpts
}

//JobCreateFactory returns a factory method for the join command
func JobCreateFactory() (cli.Command, error) {
	opts := &JobCreateOpts{}
	comm, err := newCommand("nerd job create [--name=] <image>", "Create a new compute job.", "", opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &JobCreate{
		command: comm,
		opts:    opts,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *JobCreate) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return errShowHelp("Not enough arguments, see below for usage.")
	}

	client := http.DefaultClient
	jobs, err := clusterSvc.NewJobsCaller(client, "http://localhost:3000") //@TODO make listing configurable
	if err != nil {
		return HandleError(err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*30) //@TODO allow for configuration
	defer cancel()

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}

	//@TODO handle empty names

	in := &clusterSvc.CreateJobInput{Project: ss.Project.Name, Name: cmd.opts.Name, Image: args[0]}
	out, err := jobs.CallCreate(ctx, http.MethodPost, "/jobs.create", in)
	if err != nil {
		return HandleError(err)
	}

	cmd.outputter.Logger.Printf("Created job: %s", out.GetName())
	return nil
}
