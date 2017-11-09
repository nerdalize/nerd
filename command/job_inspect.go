package command

import (
	"context"
	"net/http"
	"time"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

	clusterSvc "github.com/nerdalize/nerd/nerd/client/cluster/svc"
)

//JobInspectOpts describes command options
type JobInspectOpts struct {
	Name string `long:"name" default:"" default-mask:"" description:"Provide a unique name for this job"`
}

//JobInspect command
type JobInspect struct {
	*command
	opts *JobInspectOpts
}

//JobInspectFactory returns a factory method for the join command
func JobInspectFactory() (cli.Command, error) {
	opts := &JobInspectOpts{}
	comm, err := newCommand("nerd job list -- [cmd [args...]]", "Inspect a projects compute jobs.", "", opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &JobInspect{
		command: comm,
		opts:    opts,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *JobInspect) DoRun(args []string) (err error) {
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

	in := &clusterSvc.InspectJobInput{Project: ss.Project.Name, Name: args[0]}
	out, err := jobs.CallInspect(ctx, http.MethodGet, "/jobs.inspect", in)
	if err != nil {
		return HandleError(err)
	}

	cmd.outputter.Logger.Printf("%v", out)
	return nil
}
