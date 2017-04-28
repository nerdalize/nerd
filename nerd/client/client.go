package client

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
)

type Error struct {
	Msg        string
	Underlying error
}

func (e Error) Error() string {
	if e.Underlying != nil {
		return e.Msg + ": " + e.Underlying.Error()
	}
	return e.Msg
}

func (e Error) Cause() error {
	return e.Underlying
}

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

type Logger interface {
	Debugf(format string, args ...interface{})
	Error(args ...interface{})
}

func LogRequest(req *http.Request, logger Logger) {
	txt, err := httputil.DumpRequest(req, true)
	// retry without printing the body
	if err != nil {
		txt, err = httputil.DumpRequest(req, false)
	}
	if err == nil {
		logger.Debugf("HTTP Request:\n%s", txt)
	} else {
		logger.Error("Failed to log HTTP request")
	}
}

func LogResponse(res *http.Response, logger Logger) {
	txt, err := httputil.DumpResponse(res, true)
	// retry without printing the body
	if err != nil {
		txt, err = httputil.DumpResponse(res, false)
	}
	if err == nil {
		logger.Debugf("HTTP Response:\n%s", txt)
	} else {
		logger.Error("Failed to log HTTP response")
	}
}
