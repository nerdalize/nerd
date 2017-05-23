package command

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/olekukonko/tablewriter"
)

//WorkerListOpts describes command options
type WorkerListOpts struct {
	NerdOpts
}

//WorkerList command
type WorkerList struct {
	*command
	opts   *WorkerListOpts
	parser *flags.Parser
}

//WorkerListFactory returns a factory method for the join command
func WorkerListFactory() (cli.Command, error) {
	cmd := &WorkerList{
		command: &command{
			help:     "",
			synopsis: "show a list of all workers in the current project",
			parser:   flags.NewNamedParser("nerd worker list", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &WorkerListOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *WorkerList) DoRun(args []string) (err error) {
	config, err := conf.Read()
	if err != nil {
		HandleError(err)
	}

	bclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err)
	}

	out, err := bclient.ListWorkers(config.CurrentProject.Name)
	if err != nil {
		HandleError(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ProjectID", "WorkerID"})
	for _, t := range out.Workers {
		row := []string{}
		row = append(row, t.ProjectID)
		row = append(row, t.WorkerID)
		table.Append(row)
	}

	table.Render()
	return nil
}
