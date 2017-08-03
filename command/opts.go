package command

//OutputOpts are options that are related to CLI output.
type OutputOpts struct {
	Debug  func(bool)   `long:"debug" default:"false" optional:"true" optional-value:"true" description:"show debug output" env:"NERD_DEBUG"`
	Output func(string) `long:"output" default:"pretty" description:"format of the output" choice:"pretty" choice:"raw" choice:"json"`
}

//ConfOpts are the options related to config file and the way output is handled.
type ConfOpts struct {
	ConfigFile  func(string) `long:"config-file" default:"" default-mask:"" env:"NERD_CONFIG_FILE" description:"location of config file"`
	SessionFile func(string) `long:"session-file" default:"" default-mask:"" env:"NERD_SESSION_FILE" description:"location of session file"`
	OutputOpts
}
