package cmd_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/mitchellh/cli"

	"github.com/nerdalize/nerd/cmd"
)

func TestJobExecute(t *testing.T) {
	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	jobRun, err := cmd.JobRunFactory(ui)()
	if err != nil {
		t.Fatal("expected job run factory to return an instance, but got ", err)
	}
	command := jobRun.(*cmd.JobRun)

	err = command.Execute([]string{})
	if err == nil {
		t.Error("expected job run without arguments to return error, but got ", err)
	}
}

func TestParseInputSpecification(t *testing.T) {
	var pathTests = []struct {
		input string
		parts []string
		err   bool
	}{
		// Generic
		{"somedataset:/input", []string{"somedataset", "/input"}, false},
		{"some/relative/directory/:/input", []string{"some/relative/directory/", "/input"}, false},
		{"./dot/relative/path:/input", []string{"./dot/relative/path", "/input"}, false},
		{"./data:/~/valid/abs/path", []string{"./data", "/~/valid/abs/path"}, false},
		{"C:/input", []string{"C", "/input"}, false}, // Can we detect this "mistake"?

		// Failure cases
		{"", nil, true},
		{"nocolons", nil, true},
		{"/too:/many:/colons:/here", nil, true},
		{"./data:./relative/path", nil, true},
		{"./data:~/home/dir", nil, true},
		{"./data:", nil, true},
		{":/input", nil, true},
		{"./data:\\wrong\\separators", nil, true},

		// Windows
		{"C:/some/dir:/input", []string{"C:/some/dir", "/input"}, false},
		{"//some/dir:/input", []string{"//some/dir", "/input"}, false},
		{"C:\\some\\dir:/input", []string{"C:\\some\\dir", "/input"}, false},

		// Linux
		{"/some/abs/path:/input", []string{"/some/abs/path", "/input"}, false},
	}

	for _, testCase := range pathTests {
		parts, err := cmd.ParseInputSpecification(testCase.input)

		if testCase.err && err == nil {
			t.Errorf("expected error for input %s, but got no error", testCase.input)
		} else if !testCase.err && err != nil {
			t.Errorf("expected no error for input %s, but got %s", err)
		}

		if !reflect.DeepEqual(parts, testCase.parts) {
			t.Errorf("expected %s to be parsed into %v, but got %v", testCase.input, testCase.parts, parts)
		}
	}
}
