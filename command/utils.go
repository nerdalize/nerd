package command

import (
	"fmt"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/nerdalize/nerd/nerd/client"
	"github.com/nerdalize/nerd/nerd/payload"
	"github.com/pkg/errors"
)

type stdoutkw struct{}

func (kw *stdoutkw) Write(k string) (err error) {
	_, err = fmt.Fprintf(os.Stdout, "%v\n", k)
	return err
}

func HandleClientError(err *client.APIError, verbose bool) error {
	if perr, ok := err.Err.(*payload.Error); ok {
		var errString string
		// create error message according to response code
		switch err.Response.StatusCode {
		case 422:
			if len(perr.Fields) > 0 {
				errString = fmt.Sprintf("validation error: %v", perr.Fields)
			}
		}
		// use default server error
		if errString == "" {
			errString = perr.Error()
		}
		errString = "server side error: " + errString
		if verbose {
			errString += verboseClientError(err)
		}
		return errors.New(errString)
	} else {
		if err != nil && verbose {
			return errors.Wrap(err.Err, "debug info: \n"+verboseClientError(err))
		}
	}
	return err.Err
}

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

func HandleError(err error, verbose bool) error {
	if err != nil && verbose {
		//print stack trace
		return fmt.Errorf("%+v", err)
	}
	return err
}
