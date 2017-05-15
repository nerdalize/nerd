package command

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/olekukonko/tablewriter"
)

//QueueListOpts describes command options
type QueueListOpts struct {
	NerdOpts
}

//QueueList command
type QueueList struct {
	*command
	opts   *QueueListOpts
	parser *flags.Parser
}

//QueueListFactory returns a factory method for the join command
func QueueListFactory() (cli.Command, error) {
	cmd := &QueueList{
		command: &command{
			help:     "",
			synopsis: "show a list of all queues in the current project",
			parser:   flags.NewNamedParser("nerd queue list", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &QueueListOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return cmd, nil
}

//DoRun is called by run and allows an error to be returned
func (cmd *QueueList) DoRun(args []string) (err error) {
	config, err := conf.Read()
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	bclient, err := NewClient(cmd.ui)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	out, err := bclient.ListQueues(config.CurrentProject.Name)
	if err != nil {
		HandleError(err, cmd.opts.VerboseOutput)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ProjectID", "QueueID"})
	for _, t := range out.Queues {
		row := []string{}
		row = append(row, t.ProjectID)
		row = append(row, t.QueueID)
		table.Append(row)
	}

	table.Render()
	return nil
}
