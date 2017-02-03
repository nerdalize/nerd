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

//StatusOpts describes command options
type StatusOpts struct {
	*NerdAPIOpts
}

//Status command
type Status struct {
	*command

	ui     cli.Ui
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

	loc, err := cmd.opts.URL("/tasks")
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

	tasks := []*nerd.Task{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&tasks)
	if err != nil {
		return fmt.Errorf("failed to decode: %v", err)
	}

	for _, t := range tasks {
		fmt.Printf("%s (%s@%s): %s\n", t.ID, t.Image, t.Dataset, t.Status)
	}

	return nil
}
