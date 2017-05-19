package command

//OutputOpts are options that are related to CLI output.
type OutputOpts struct {
	VerboseOutput func(bool) `short:"v" long:"verbose" default:"false" optional:"true" optional-value:"true" description:"show verbose output"`
	JSONOutput    func(bool) `long:"json-format" default:"false" optional:"true" optional-value:"true" description:"show output in json format"`
}

//NerdOpts are the options that are applicable to all nerd commands.
type NerdOpts struct {
	ConfOpts
}

//ConfOpts are the options related to config file and the way output is handled.
type ConfOpts struct {
	ConfigFile func(string) `long:"config" default:"" default-mask:"" env:"CONFIG" description:"location of config file"`
	OutputOpts
}
