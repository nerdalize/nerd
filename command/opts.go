package command

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/nerdalize/nerd/nerd/client"
)

//NerdAPIOpts configure how the platform endpoint is reached
type NerdAPIOpts struct {
	NerdAPIScheme string `long:"api-scheme" default:"https" default-mask:"https" env:"NERD_API_SCHEME" description:"protocol for endpoint communication"`

	NerdAPIHostname string `long:"api-hostname" default:"platform.nerdalize.net" default-mask:"platform.nerdalize.net" env:"NERD_API_HOST" description:"hostname of the compute platform"`

	NerdAPIBasePath string `long:"api-basepath" default:"" default-mask:"" env:"NERD_API_BASE_PATH" description:"basepath of the endpoint"`

	NerdAPIVersion string `long:"api-version" default:"v1" default-mask:"v1" env:"NERD_API_VERSION" description:"endpoint version"`

	NerdAPIFullURL string `long:"api-full-url" default:"" default-mask:"" env:"NERD_API_FULL_URL" description:"full endpoint url"`
}

//URL returns a fully qualitied url on the platform endpoint
func (opts *NerdAPIOpts) URL(path string) (loc *url.URL, err error) {
	loc, err = url.Parse(fmt.Sprintf(
		"%s://%s/%s/%s",
		opts.NerdAPIScheme,
		opts.NerdAPIHostname,
		opts.NerdAPIVersion,
		strings.TrimLeft(path, "/"),
	))
	return loc, err
}

//NerdAPIConfig returns a populated configuration struct for the Nerd API client.
func (opts *NerdAPIOpts) NerdAPIConfig() client.NerdAPIConfig {
	return client.NerdAPIConfig{
		Scheme:   opts.NerdAPIScheme,
		Host:     opts.NerdAPIHostname,
		BasePath: opts.NerdAPIBasePath,
		Version:  opts.NerdAPIVersion,
	}
}
