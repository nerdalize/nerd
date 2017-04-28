package v1batch

import (
	"fmt"
	"net/http"

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
	switch e.StatusCode {
	case http.StatusUnprocessableEntity:
		if len(e.Err.Fields) > 0 {
			return fmt.Sprintf("Validation error: %v", e.Err.Fields)
		}
	case http.StatusNotFound:
		return fmt.Sprint("The specified resource does not exist")
	}
	return fmt.Sprintf("unknown server error (%v)", e.StatusCode)
}

//UserFacingMsg is implemented to make sure this message is shown to an end user.
func (e HTTPError) UserFacingMsg() string {
	return e.Error()
}

//Underlying is part of the user facing interface.
func (e HTTPError) Underlying() error {
	return nil
}

//Cause is implemented to be compatible with the pkg/errors package.
func (e HTTPError) Cause() error {
	return nil
}
