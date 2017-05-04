package command

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	nerdaws "github.com/nerdalize/nerd/nerd/aws"
	"github.com/nerdalize/nerd/nerd/client/batch/v1"
	"github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/pkg/errors"
)

//WorkerWorkOpts describes command options
type WorkerWorkOpts struct {
	NerdOpts
}

//WorkerWork command
type WorkerWork struct {
	*command
	opts   *WorkerWorkOpts
	parser *flags.Parser
}

//WorkerWorkFactory returns a factory method for the join command
func WorkerWorkFactory() (cli.Command, error) {
	cmd := &WorkerWork{
		command: &command{
			help:     "",
			synopsis: "start working tasks of a queue locally",
			parser:   flags.NewNamedParser("nerd worker work <queue-id> <command-tmpl> [arg-tmpl...]", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &WorkerWorkOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkerWork) DoRun(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	config, err := conf.Read()
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	bclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	creds := nerdaws.NewNerdalizeCredentials(bclient, config.CurrentProject)
	qops, err := nerdaws.NewQueueClient(creds, "eu-west-1")
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	targs := []*template.Template{}
	for i, arg := range args[1:] {
		var t *template.Template
		t, err = template.New(fmt.Sprintf("arg%d", i)).Parse(arg)
		if err != nil {
			HandleError(errors.Wrap(err, "failed to parse argument as template"), cmd.opts.VerboseOutput)
		}

		targs = append(targs, t)
	}

	//receiving task runs
	runCh := make(chan *v1payload.Run)
	go func() {
		for {
			var out []*v1payload.Run
			out, err = bclient.ReceiveTaskRuns(config.CurrentProject, args[0], time.Second*20, qops)
			if err != nil {
				HandleError(err, cmd.opts.VerboseOutput)
				close(runCh)
				return
			}

			for _, run := range out {
				runCh <- run
			}
		}
	}()

	//main event loop
	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case run := <-runCh: //new run
			if run == nil {
				return fmt.Errorf("run receive loop exited, unexpectedly")
			}

			go func() {
				err = doRun(bclient, targs, run)
				if err != nil {
					logrus.Errorf("failing run of task %d: %v", run.TaskID, err)
				} else {
					logrus.Infof("run of task %d succeeded", run.TaskID)
				}
			}()
		case sig := <-exitCh: //exit signal
			logrus.Infof("received signal %s, exiting...", sig)
			return
		}
	}
}

func doRun(bclient *v1batch.Client, targs []*template.Template, r *v1payload.Run) (err error) {
	defer func() {
		if err != nil {
			//@TODO find a better way to check if the process exited by manual killing
			if !strings.Contains(err.Error(), "killed") {
				_, ferr := bclient.SendRunFailure(r.ProjectID, r.QueueID, r.TaskID, r.Token, "MY-CODE", "MY MESSAGE")
				if ferr != nil {
					logrus.Errorf("failed to send run failure: %v", ferr)
				}
			}

		} else {
			//@TODO how do we send over results
			_, serr := bclient.SendRunSuccess(r.ProjectID, r.QueueID, r.TaskID, r.Token, "MY-SUCCESS")
			if serr != nil {
				logrus.Errorf("failed to send run success: %v", serr)
			}
		}
	}()

	pl := map[string]interface{}{}
	err = json.Unmarshal([]byte(r.Payload), &pl)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal payload as JSON map")
	}

	args := []string{}
	for _, t := range targs {
		buf := bytes.NewBuffer(nil)
		err = t.Execute(buf, pl)
		if err != nil {
			return errors.Wrap(err, "failed to execute templated arguments")
		}

		args = append(args, buf.String())
	}

	if len(args) < 1 {
		return errors.Errorf("templated args need at least on element to have something to run")
	}

	ppath, err := exec.LookPath(args[0])
	if err != nil {
		return errors.Errorf("couldn't find executable named '%s', current PATH: %v", args[0], os.Getenv("PATH"))
	}

	stopProcessCh := make(chan struct{})
	stopHeartbeatCh := make(chan struct{})
	go func() {
		tick := time.Tick(time.Second * 10)
		for {
			select {
			case <-tick:
				var out *v1payload.SendRunHeartbeatOutput
				out, err = bclient.SendRunHeartbeat(r.ProjectID, r.QueueID, r.TaskID, r.Token)
				logrus.Infof("heartbeat %d out: %+v", r.TaskID, out)
				if err != nil || out == nil || out.HasExpired {
					close(stopProcessCh)
					return
				}

			case <-stopHeartbeatCh:
				return
			}
		}
	}()

	logrus.Infof("starting run of task %d: %s %v", r.TaskID, ppath, args[1:])
	cmd := exec.Command(ppath, args[1:]...)
	cmd.Stdout = os.Stderr //@TODO discard? send to logging
	cmd.Stderr = os.Stderr //@TODO send logging
	err = cmd.Start()
	if err != nil {
		return errors.Wrap(err, "failed to start run")
	}

	go func() {
		<-stopProcessCh
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	err = cmd.Wait()
	close(stopHeartbeatCh)
	if err != nil {
		return errors.Wrap(err, "waiting on process failed")
	}

	if cmd.ProcessState == nil {
		return errors.Errorf("no process state")
	}

	if !cmd.ProcessState.Success() {
		return errors.Errorf("run failed: %+v", cmd.ProcessState)
	}

	return nil
}
