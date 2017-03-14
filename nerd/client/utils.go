package client

import (
	"net/http"
	"net/http/httputil"

	"github.com/Sirupsen/logrus"
)

func logRequest(req *http.Request) {
	txt, err := httputil.DumpRequest(req, true)
	// retry without printing the body
	if err != nil {
		txt, err = httputil.DumpRequest(req, false)
	}
	if err == nil {
		logrus.Debugf("HTTP Request:\n%s", txt)
	} else {
		logrus.Error("Failed to log HTTP request")
	}
}

func logResponse(res *http.Response) {
	txt, err := httputil.DumpResponse(res, true)
	// retry without printing the body
	if err != nil {
		txt, err = httputil.DumpResponse(res, false)
	}
	if err == nil {
		logrus.Debugf("HTTP Response:\n%s", txt)
	} else {
		logrus.Error("Failed to log HTTP response")
	}
}
