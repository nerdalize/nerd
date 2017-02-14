package client

import "net/http"

//APIError is the error that is returned by client.NerdAPIClient
type APIError struct {
	Request  *http.Request
	Response *http.Response
	Err      error
}

func (e APIError) Error() string {
	return e.Err.Error()
}

func (e APIError) Cause() error {
	return e.Err
}
