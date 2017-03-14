package command

//OutputOpts are options that are related to CLI output.
type OutputOpts struct {
	VerboseOutput bool `short:"v" long:"verbose" default-mask:"false" description:"show verbose output"`
	JSONOutput    bool `long:"json-format" default-mask:"false" description:"show output in json format"`
}

//NerdOpts are the options that are applicable to all nerd commands.
type NerdOpts struct {
	OutputOpts
}
