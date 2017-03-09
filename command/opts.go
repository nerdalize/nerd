package command

//NerdAPIOpts configure how the platform endpoint is reached
// TODO: Change to one flag for URL
type NerdAPIOpts struct {
	// NerdAPIScheme string `long:"api-scheme" default:"https" default-mask:"https" env:"NERD_API_SCHEME" description:"protocol for endpoint communication"`
	//
	// NerdAPIHostname string `long:"api-hostname" default:"platform.nerdalize.net" default-mask:"platform.nerdalize.net" env:"NERD_API_HOST" description:"hostname of the compute platform"`
	//
	// NerdAPIBasePath string `long:"api-basepath" default:"" default-mask:"" env:"NERD_API_BASE_PATH" description:"basepath of the endpoint"`
	//
	// NerdAPIVersion string `long:"api-version" default:"v1" default-mask:"v1" env:"NERD_API_VERSION" description:"endpoint version"`
	//
	// NerdAPIURL string `long:"api-url" default:"" default-mask:"" env:"NERD_API_URL" description:"full endpoint url"`
}

//ConfOpts is used to set the location of the config file.
type ConfOpts struct {
	ConfigFile string `long:"config" default:"" default-mask:"" env:"CONFIG" description:"location of config file"`
}

//OutputOpts is used to determine how output should be presented to the user.
type OutputOpts struct {
	VerboseOutput bool `short:"v" long:"verbose" default-mask:"false" description:"show verbose output"`
}

//NerdOpts are the options that are applicable to all nerd commands.
type NerdOpts struct {
	NerdAPIOpts
	OutputOpts
	ConfOpts
}

//URL returns a fully qualitied url on the platform endpoint
// func (opts *NerdAPIOpts) URL() (loc string) {
// 	if opts.NerdAPIURL != "" {
// 		return opts.NerdAPIURL
// 	}
// 	return fmt.Sprintf(
// 		"%s://%s/%s/%s",
// 		opts.NerdAPIScheme,
// 		opts.NerdAPIHostname,
// 		opts.NerdAPIVersion,
// 	)
// }

//NerdAPIConfig returns a populated configuration struct for the Nerd API client.
// func (opts *NerdAPIOpts) NerdAPIConfig() client.NerdAPIConfig {
// 	return client.NerdAPIConfig{
// 		Scheme:   opts.NerdAPIScheme,
// 		Host:     opts.NerdAPIHostname,
// 		BasePath: opts.NerdAPIBasePath,
// 		Version:  opts.NerdAPIVersion,
// 	}
// }
