package client

import (
	"fmt"
	"net/http"

	"github.com/nerdalize/nerd/nerd/payload"
)

//HTTPError encapsulates an error over the HTTP protocol
type HTTPError struct {
	StatusCode int
	Err        *payload.Error
}

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

//UserFacingMsg is a user friendly message
func (e HTTPError) UserFacingMsg() string {
	return e.Error()
}

//Underlying is the system error
func (e HTTPError) Underlying() error {
	return nil
}
