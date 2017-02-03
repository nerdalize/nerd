package command

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd"
)

//LogsOpts describes command options
type LogsOpts struct {
	*NerdAPIOpts
}

//Logs command
type Logs struct {
	*command

	ui     cli.Ui
	opts   *LogsOpts
	parser *flags.Parser
}

//LogsFactory returns a factory method for the join command
func LogsFactory() func() (cmd cli.Command, err error) {
	cmd := &Logs{
		command: &command{
			help:     "",
			synopsis: "retrieve up-to-date feedback from a task",
			parser:   flags.NewNamedParser("nerd logs <task_id>", flags.Default),
			ui: &cli.BasicUi{
				Reader: os.Stdin,
				Writer: os.Stderr,
			},
		},

		opts: &LogsOpts{},
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
func (cmd *Logs) DoRun(args []string) (err error) {
	if len(args) < 1 {
		return fmt.Errorf("not enough arguments, see --help")
	}

	loc, err := cmd.opts.URL("/tasks/" + args[0])
	if err != nil {
		return fmt.Errorf("failed to create API url from cli options: %+v", err)
	}

	req, err := http.NewRequest("GET", loc.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create API request: %+v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("API request '%s %s' failed: %v", req.Method, loc, err)
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API request '%s %s' returned unexpected status from API: %v", req.Method, loc, resp.Status)
	}

	t := &nerd.Task{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(t)
	if err != nil {
		return fmt.Errorf("failed to deserialize: %v", err)
	}

	for _, line := range t.LogLines {
		fmt.Println(line)
	}

	return nil
}
