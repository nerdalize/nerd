package v2client

import (
	"net/http"
	"net/http/httputil"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Error(args ...interface{})
}

func logRequest(req *http.Request, logger Logger) {
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

func logResponse(res *http.Response, logger Logger) {
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
