package main_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/command"
)

func TestMain(m *testing.M) {
	if os.Getenv("NERD_API_TOKEN") == "" {
		fmt.Println("no NERD_API_TOKEN set in environment")
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestUploadCommand(t *testing.T) {
	ui := &cli.MockUi{}
	cmdf := command.UploadFactory(ui) //provide ui as arg to the factory
	cmd, err := cmdf()
	ok(t, err)

	exitc := cmd.Run([]string{})
	equals(t, 1, exitc)
	assertbuf(t, ui.ErrorWriter, "not enough arguments")
	assertbuf(t, ui.OutputWriter, ".*")
}
