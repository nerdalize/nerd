package client

import "net/http"

type APIError struct {
	Request  *http.Request
	Response *http.Response
	Err      error
}

func (e APIError) Error() string {
	return e.Err.Error()
}
