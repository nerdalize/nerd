package command

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/nerdalize/nerd/nerd/client"
	"github.com/nerdalize/nerd/nerd/client/credentials/provider"
	"github.com/nerdalize/nerd/nerd/payload"
	"github.com/pkg/errors"
)

const (
	debugHeader = "\n\n[DEBUG INFO]:"
)

type stdoutkw struct{}

func (kw *stdoutkw) Write(k string) (err error) {
	_, err = fmt.Fprintf(os.Stdout, "%v\n", k)
	return err
}

//NewClient creates a new NerdAPIClient with two credential providers.
func NewClient(ui cli.Ui, nerdAPIURL, authURL string) *client.NerdAPIClient {
	return client.NewNerdAPIWithEndpoint(provider.NewChainCredentials(
		provider.NewEnvDisk(),
		provider.NewAuthAPI(func() (string, string, error) {
			ui.Info("Please enter your Nerdalize username and password.")
			user, err := ui.Ask("Username: ")
			if err != nil {
				return "", "", errors.Wrap(err, "failed to read username")
			}
			pass, err := ui.AskSecret("Password: ")
			if err != nil {
				return "", "", errors.Wrap(err, "failed to read password")
			}
			return user, pass, nil
		}, client.NewAuthAPI(authURL)),
	), nerdAPIURL)
}

//HandleClientError handles errors produced by client.NerdAPIClient
func HandleClientError(err error, verbose bool) error {
	// only handle *client.APIError
	aerr, ok := err.(*client.APIError)
	if !ok {
		return err
	}
	ret := aerr.Err
	if perr, ok := aerr.Err.(*payload.Error); ok && aerr.Response != nil {
		// create error message according to response code
		switch aerr.Response.StatusCode {
		case http.StatusUnprocessableEntity:
			if len(perr.Fields) > 0 {
				ret = errors.Wrapf(perr, "validation error: %v", perr.Fields)
			}
		}
	}
	if verbose {
		return errors.Wrap(ret, debugHeader+verboseClientError(aerr))
	}
	return ret
}

//verboseClientError creates pretty formatted represntations of HTTP request and response.
func verboseClientError(aerr *client.APIError) string {
	var message []string

	if aerr.Request != nil {
		message = append(message, "", "HTTP Request:")

		req, err := httputil.DumpRequest(aerr.Request, true)
		// retry without printing the body
		if err != nil {
			req, err = httputil.DumpRequest(aerr.Request, false)
		}
		if err == nil {
			message = append(message, string(req))
		}
	}

	if aerr.Response != nil {
		message = append(message, "", "HTTP Response:")
		resp, err := httputil.DumpResponse(aerr.Response, true)
		// retry without printing the body
		if err != nil {
			resp, err = httputil.DumpResponse(aerr.Response, false)
		}
		if err == nil {
			message = append(message, string(resp))
		}
	}

	return strings.Join(message, "\n")
}

//HandleError handles the way errors are presented to the user.
func HandleError(err error, verbose bool) error {
	if verbose {
		return fmt.Errorf("%+v", err)
	}
	// when there's are more than 1 message on the message stack, only print the top one for user friendlyness.
	//TODO this will return the entire error chain except the last error. Created issue on GH: https://github.com/pkg/errors/issues/104
	if errors.Cause(err) != nil {
		fmt.Println(errors.Cause(err))
		return fmt.Errorf(strings.Replace(err.Error(), ": "+errors.Cause(err).Error(), "", 1))
	}
	return err
}
