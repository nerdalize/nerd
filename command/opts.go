package command

//OutputOpts are options that are related to CLI output.
type OutputOpts struct {
	VerboseOutput func(bool)   `short:"v" long:"verbose" default:"false" optional:"true" optional-value:"true" description:"show verbose output"`
	JSONOutput    func(bool)   `long:"json-format" default:"false" optional:"true" optional-value:"true" description:"show output in json format"`
	Output        func(string) `long:"output" default:"text" description:"[text|json]"`
}

//ConfOpts are the options related to config file and the way output is handled.
type ConfOpts struct {
	ConfigFile  func(string) `long:"config-file" default:"" default-mask:"" env:"NERD_CONFIG_FILE" description:"location of config file"`
	SessionFile func(string) `long:"session-file" default:"" default-mask:"" env:"NERD_SESSION_FILE" description:"location of session file"`
	OutputOpts
}
