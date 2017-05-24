package command

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/cli"
	"github.com/pkg/errors"
)

//ProjectPlace command
type ProjectPlace struct {
	*command
}

//ProjectPlaceFactory returns a factory method for the join command
func ProjectPlaceFactory() (cli.Command, error) {
	comm, err := newCommand("nerd project place <host> <token>", "place the current project on a compute cluster", "", nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create command")
	}
	cmd := &ProjectPlace{
		command: comm,
	}

	cmd.runFunc = cmd.DoRun
	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *ProjectPlace) DoRun(args []string) (err error) {
	if len(args) < 2 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	bclient, err := NewClient(cmd.ui, cmd.config, cmd.session)
	if err != nil {
		HandleError(err)
	}

	ss, err := cmd.session.Read()
	if err != nil {
		HandleError(err)
	}

	out, err := bclient.PlaceProject(ss.Project.Name, args[0], args[1], "") //@TODO allow self signed certificate
	if err != nil {
		HandleError(err)
	}

	logrus.Infof("Placement created: %v", out)
	return nil
}
