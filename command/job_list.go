package command

import (
	"context"
	"net/http"
	"time"

	"github.com/mitchellh/cli"
	"github.com/pkg/errors"

	"github.com/nerdalize/nerd/command/format"
	clusterSvc "github.com/nerdalize/nerd/nerd/client/cluster/svc"
)

//JobListOpts describes command options
type JobListOpts struct{}

//JobList command
type JobList struct {
	*command
	opts *JobListOpts
}

//JobListFactory returns a factory method for the join command
func JobListFactory() (cli.Command, error) {
	opts := &JobListOpts{}
	comm, err := newCommand("nerd job list -- [cmd [args...]]", "List a projects compute jobs.", "", opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &JobList{
		command: comm,
		opts:    opts,
	}
	cmd.runFunc = cmd.DoRun

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *JobList) DoRun(args []string) (err error) {
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

	in := &clusterSvc.ListJobsInput{Project: ss.Project.Name}
	out, err := jobs.CallList(ctx, http.MethodGet, "/jobs.list", in)
	if err != nil {
		return HandleError(err)
	}

	header := "JOB\tSTATUS\tCREATION TIME"
	pretty := "{{range $i, $x := $.Results}}{{$x.Name}}\t{{$x.Status}}\t{{$x.CreatedAt}}\n{{end}}"
	raw := "{{range $i, $x := $.Results}}{{$x.Name}}\t{{$x.Status}}\t{{$x.CreatedAt}}\n{{end}}"
	cmd.outputter.Output(format.DecMap{
		format.OutputTypePretty: format.NewTableDecorator(out, header, pretty),
		format.OutputTypeRaw:    format.NewTmplDecorator(out, raw),
		format.OutputTypeJSON:   format.NewJSONDecorator(out),
	})

	return nil
}
