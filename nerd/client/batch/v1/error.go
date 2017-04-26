package v1batch

import (
	"fmt"
	"net/http"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

type HTTPError struct {
	StatusCode int
	Err        *v1payload.Error
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

func (e HTTPError) UserFacingMsg() string {
	return e.Error()
}

func (e HTTPError) Underlying() error {
	return nil
}

func (e HTTPError) Cause() error {
	return nil
}
