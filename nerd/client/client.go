package client

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	log "github.com/sirupsen/logrus"
)

//Error is the default error returned by clients in the client package.
//Error is compatible with the pkg/errors package.
type Error struct {
	Msg        string
	Underlying error
}

//NewError creates a new Error
func NewError(msg string, underlying error) *Error {
	return &Error{
		Msg:        msg,
		Underlying: underlying,
	}
}

//Error returns the error message.
func (e Error) Error() string {
	if e.Underlying != nil {
		return e.Msg + ": " + e.Underlying.Error()
	}
	return e.Msg
}

//Cause points to the underlying error.
func (e Error) Cause() error {
	return e.Underlying
}

//Format implements different error formats.
func (e Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if e.Underlying != nil {
				fmt.Fprintf(s, "%+v\n", e.Underlying)
			}
			io.WriteString(s, e.Msg)
			return
		}
		fallthrough
	case 's', 'q':
		io.WriteString(s, e.Error())
	}
}

//LogRequest is a util to log an HTTP request.
func LogRequest(req *http.Request, logger *log.Logger) {
	txt, err := httputil.DumpRequest(req, true)
	// retry without printing the body
	if err != nil {
		txt, err = httputil.DumpRequest(req, false)
	}
	if err == nil {
		logger.Printf("[DEBUG] HTTP Request:\n%s\n", txt)
	} else {
		logger.Printf("[DEBUG] Failed to log HTTP request '%v'", err)
	}
}

//LogResponse is a util to log an HTTP response.
func LogResponse(res *http.Response, logger *log.Logger) {
	txt, err := httputil.DumpResponse(res, true)
	// retry without printing the body
	if err != nil {
		txt, err = httputil.DumpResponse(res, false)
	}
	if err == nil {
		logger.Printf("[DEBUG] HTTP Response:\n%s\n", txt)
	} else {
		logger.Printf("[DEBUG] Failed to log HTTP response '%v'", err)
	}
}
