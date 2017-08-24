package command

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/jessevdk/go-flags"
	"github.com/mitchellh/cli"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/nerdalize/nerd/command/format"
	"github.com/nerdalize/nerd/nerd/conf"
)

var (
	//SharedOptionsGroup are options shared by all commands
	SharedOptionsGroup = "Output options"
)

const (
	//EnvConfigJSON can be used to pass the config file as a json encoded string
	EnvConfigJSON = "NERD_CONFIG_JSON"

	//EnvNerdProject can be used to set the nerd project
	EnvNerdProject = "NERD_PROJECT"
)

// errShowHelp is used to retrieve the cause of an error.
// If the cause type is errShowHelp, this will print the help for the (sub) command
type errShowHelp string

func (e errShowHelp) Error() string { return string(e) }

func newCommand(usage, synopsis, help string, opts interface{}) (*command, error) {
	cmd := &command{
		usage:    usage,
		help:     help,
		synopsis: synopsis,
		parser:   flags.NewNamedParser(usage, flags.Default),
		ui: &cli.BasicUi{
			Reader: os.Stdin,
			Writer: os.Stderr,
		},
		outputter: format.NewOutputter(os.Stdout, os.Stderr, log.New(os.Stderr, "", 0)),
	}
	if opts != nil {
		_, err := cmd.parser.AddGroup("options", "options", opts)
		if err != nil {
			return nil, err
		}
	}
	confOpts := &ConfOpts{
		ConfigFile:  cmd.setConfig,
		SessionFile: cmd.setSession,
		OutputOpts: OutputOpts{
			Output: cmd.setOutput,
			Debug:  cmd.setDebug,
		},
	}
	_, err := cmd.parser.AddGroup(SharedOptionsGroup, SharedOptionsGroup, confOpts)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

//command is an abstract implementation for embedding in concrete commands and allows basic command functionality to be reused.
type command struct {
	usage     string
	help      string        //extended help message, show when --help a command
	synopsis  string        //short help message, shown on the command overview
	parser    *flags.Parser //option parser that will be used when parsing args
	ui        cli.Ui
	config    *conf.Config
	outputter *format.Outputter
	session   *conf.Session
	runFunc   func(args []string) error
}

func (c *command) Options() *flags.Parser {
	return c.parser
}

func (c *command) Usage() string {
	return c.usage
}

func (c *command) Description() string {
	txt := c.help
	if txt == "" {
		txt = c.Synopsis()
	}
	return txt
}

//Will write help text for when a user uses --help, it automatically renders all option groups of the flags.Parser (augmented with default values). It will show an extended help message if it is not empty, else it shows the synopsis.
func (c *command) Help() string {
	buf := bytes.NewBuffer(nil)
	c.parser.WriteHelp(buf)
	return fmt.Sprintf(`
%s

%s`, c.Description(), buf.String())
}

//Short explanation of the command as passed in the struction initialization
func (c *command) Synopsis() string {
	return c.synopsis
}

//Run wraps a signature that allows returning an error type and parses the arguments for the flags package. If flag parsing fails it sets the exit code to 127, if the command implementation returns a non-nil error the exit code is 1
func (c *command) Run(args []string) int {
	if c.parser != nil {
		var err error
		args, err = c.parser.ParseArgs(args)
		if err != nil {
			return 127
		}
	}

	if err := c.runFunc(args); err != nil {
		switch cause := errors.Cause(err).(type) {
		case errShowHelp:
			c.outputter.WriteError(cause)
			return cli.RunResultHelp
		default:
			c.outputter.WriteError(err)
			return 1
		}
	}

	return 0
}

//setConfig sets the cmd.config field according to the config file location
func (c *command) setConfig(loc string) {
	if json := os.Getenv(EnvConfigJSON); json != "" {
		conf, err := conf.FromJSON(json)
		if err != nil {
			fmt.Fprint(os.Stderr, errors.Wrapf(err, "failed to parse config json '%v'", json))
			os.Exit(-1)
		}
		c.config = conf
		return
	}
	if loc == "" {
		var err error
		loc, err = conf.GetDefaultConfigLocation()
		if err != nil {
			c.outputter.WriteError(errors.Wrap(err, "failed to find config location"))
			os.Exit(-1)
		}
		err = createFile(loc, "{}")
		if err != nil {
			c.outputter.WriteError(errors.Wrapf(err, "failed to create config file %v", loc))
			os.Exit(-1)
		}
	}
	conf, err := conf.Read(loc)
	if err != nil {
		c.outputter.WriteError(errors.Wrap(err, "failed to read config file"))
		os.Exit(-1)
	}
	c.config = conf
	if conf.Logging.Enabled {
		logPath, err := homedir.Expand(conf.Logging.FileLocation)
		if err != nil {
			c.outputter.WriteError(errors.Wrap(err, "failed to find home directory"))
			os.Exit(-1)
		}
		err = createFile(logPath, "")
		if err != nil {
			c.outputter.WriteError(errors.Wrapf(err, "failed to create log file %v", logPath))
			os.Exit(-1)
		}
		err = c.outputter.SetLogToDisk(logPath)
		if err != nil {
			c.outputter.WriteError(errors.Wrap(err, "failed to set logging"))
			os.Exit(-1)
		}
	}
}

//setSession sets the cmd.session field according to the session file location
func (c *command) setSession(loc string) {
	if loc == "" {
		var err error
		loc, err = conf.GetDefaultSessionLocation()
		if err != nil {
			c.outputter.WriteError(errors.Wrap(err, "failed to find session location"))
			os.Exit(-1)
		}
		err = createFile(loc, "{}")
		if err != nil {
			c.outputter.WriteError(errors.Wrapf(err, "failed to create session file %v", loc))
			os.Exit(-1)
		}
	}
	c.session = conf.NewSession(loc)
	if proj := os.Getenv(EnvNerdProject); proj != "" {
		c.session.WriteProject(proj, conf.DefaultAWSRegion)
	}
}

//setDebug sets debug output formatting
func (c *command) setDebug(debug bool) {
	c.outputter.SetDebug(debug)
}

//setOutput specifies the type of output
func (c *command) setOutput(output string) {
	switch output {
	case "json":
		c.outputter.SetOutputType(format.OutputTypeJSON)
	case "raw":
		c.outputter.SetOutputType(format.OutputTypeRaw)
	case "pretty":
		fallthrough
	default:
		c.outputter.SetOutputType(format.OutputTypePretty)
	}
}

func createFile(path, content string) error {
	os.MkdirAll(filepath.Dir(path), 0755)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil && !os.IsExist(err) {
		return err
	}
	if err == nil {
		f.Write([]byte(content))
	}
	f.Close()
	return nil
}
