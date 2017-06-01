package command

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//ProjectPlaceOpts describes command options
type ProjectPlaceOpts struct {
	Token    string `long:"token" default:"" default-mask:"" description:"placement that authenticates using JWT"`
	Username string `long:"username" default:"" default-mask:"" description:"username for placement that authenticates using username/password"`
	Password string `long:"password" default:"" default-mask:"" description:"password for placement that authenticates using username/password"`
	Insecure bool   `long:"insecure" default-mask:"" description:"disable checking of server certificate"`
}

//ProjectPlace command
type ProjectPlace struct {
	*command
	opts *ProjectPlaceOpts
}

//ProjectPlaceFactory returns a factory method for the join command
func ProjectPlaceFactory() (cli.Command, error) {
	opts := &ProjectPlaceOpts{}
	comm, err := newCommand("nerd project place <host>", "place the current project on a compute cluster", "", opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &ProjectPlace{
		command: comm,
		opts:    opts,
	}

	cmd.runFunc = cmd.DoRun
	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *ProjectPlace) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	bclient, err := NewClient(cmd.config, cmd.session, cmd.outputter)
	if err != nil {
		return HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		return HandleError(err)
	}

	host := args[0]
	token := cmd.opts.Token
	username := cmd.opts.Username
	password := cmd.opts.Password
	insecure := cmd.opts.Insecure

	out, err := bclient.PlaceProject(ss.Project.Name, host, token, "", username, password, insecure)
	if err != nil {
		return HandleError(err)
	}

	logrus.Infof("Placement created: %v", out)
	return nil
}
