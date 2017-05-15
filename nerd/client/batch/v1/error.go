package v1batch

import (
	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

//HTTPError is an error that is used when a server responded with a status code >= 400.
//Based on the actual status code a custom error message will be generated.
type HTTPError struct {
	StatusCode int
	Err        *v1payload.Error
}

//Error returns the error message specific for the status code.
func (e HTTPError) Error() string {
	return e.Err.Message
}

//Cause is implemented to be compatible with the pkg/errors package.
func (e HTTPError) Cause() error {
	return nil
}
