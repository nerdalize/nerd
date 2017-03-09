package command

import (
	"os"

	humanize "github.com/dustin/go-humanize"
	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/conf"
	"github.com/nerdalize/nerd/nerd/payload"
	"github.com/olekukonko/tablewriter"
)

//StatusOpts describes command options
type StatusOpts struct {
	NerdOpts
}

//Status command
type Status struct {
	*command

	opts   *StatusOpts
	parser *flags.Parser
}

//StatusFactory returns a factory method for the join command
func StatusFactory() func() (cmd cli.Command, err error) {
	cmd := &Status{
		command: &command{
			help:     "",
			synopsis: "show the status of all queued tasks",
			parser:   flags.NewNamedParser("nerd status", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &StatusOpts{},
	}

	cmd.runFunc = cmd.DoRun
	_, err := cmd.command.parser.AddGroup("options", "options", cmd.opts)
	if err != nil {
		panic(err)
	}

	return func() (cli.Command, error) {
		return cmd, nil
	}
}

//DoRun is called by run and allows an error to be returned
func (cmd *Status) DoRun(args []string) (err error) {
	conf.SetLocation(cmd.opts.ConfigFile)

	client, err := NewClient(cmd.ui)
	if err != nil {
		return HandleError(HandleClientError(err, cmd.opts.VerboseOutput), cmd.opts.VerboseOutput)
	}
	tasks, err := client.ListTasks()
	if err != nil {
		return HandleError(HandleClientError(err, cmd.opts.VerboseOutput), cmd.opts.VerboseOutput)
	}

	drawTable(tasks)

	return nil
}

func drawTable(tasks *payload.TaskListOutput) {
	data := make([][]string, len(tasks.Tasks))
	for i, task := range tasks.Tasks {
		data[i] = []string{
			task.TaskID,
			task.OutputID,
			humanize.Time(task.CreatedAt),
		}
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"TaskID", "Output Dataset", "Created"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")
	table.AppendBulk(data) // Add Bulk Data
	table.Render()
}
