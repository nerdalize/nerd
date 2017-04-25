package v1batch

import (
	"fmt"
	"io"
	"net/http"

	v2payload "github.com/nerdalize/nerd/nerd/payload/v2"
)

type Error struct {
	msg        string
	underlying error
}

func (e Error) Error() string {
	if e.underlying != nil {
		return e.msg + ": " + e.underlying.Error()
	}
	return e.msg
}

func (e Error) Cause() error {
	return e.underlying
}

func (e Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if e.underlying != nil {
				fmt.Fprintf(s, "%+v\n", e.underlying)
			}
			io.WriteString(s, e.msg)
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, e.Error())
	}
}

type HTTPError struct {
	StatusCode int
	Err        *v2payload.Error
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
