package command

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/nerdalize/nerd/nerd/client"
	"github.com/nerdalize/nerd/nerd/payload"
	"github.com/pkg/errors"
)

func newRequest() *http.Request {
	req, _ := http.NewRequest("GET", "http://nerdalize.com", nil)
	return req
}

func newResponse(statuscode int) *http.Response {
	return &http.Response{
		StatusCode: statuscode,
	}
}

func newAPIError(statuscode int) *client.APIError {
	return &client.APIError{
		Request:  newRequest(),
		Response: newResponse(statuscode),
	}
}

func TestHandleClientError(t *testing.T) {
	tests := map[string]struct {
		err      error
		verbose  bool
		expected error
	}{
		"nil error": {
			err:      nil,
			verbose:  false,
			expected: nil,
		},
		"nil error verbose": {
			err:      nil,
			verbose:  true,
			expected: nil,
		},
		"invalid error": {
			err:      fmt.Errorf("this error is not of type *client.APIError"),
			verbose:  false,
			expected: fmt.Errorf("this error is not of type *client.APIError"),
		},
		"invalid error verbose": {
			err:      fmt.Errorf("this error is not of type *client.APIError"),
			verbose:  true,
			expected: fmt.Errorf("this error is not of type *client.APIError"),
		},
		"api error": {
			err: &client.APIError{
				Request:  nil,
				Response: nil,
				Err:      fmt.Errorf("this error is not of type *payload.Error"),
			},
			verbose:  false,
			expected: fmt.Errorf("this error is not of type *payload.Error"),
		},
		"api error verbose": {
			err: &client.APIError{
				Request:  nil,
				Response: nil,
				Err:      fmt.Errorf("this error is not of type *payload.Error"),
			},
			verbose:  true,
			expected: fmt.Errorf(debugHeader + ": this error is not of type *payload.Error"),
		},
		"payload error": {
			err: &client.APIError{
				Request:  newRequest(),
				Response: newResponse(http.StatusUnprocessableEntity),
				Err: &payload.Error{
					Message: "payload error",
					Fields: map[string]string{
						"foo": "bar",
					},
				},
			},
			verbose:  false,
			expected: fmt.Errorf("validation error: %v: payload error", map[string]string{"foo": "bar"}),
		},
		"payload error verbose": {
			err: &client.APIError{
				Request:  newRequest(),
				Response: newResponse(http.StatusUnprocessableEntity),
				Err: &payload.Error{
					Message: "payload error",
					Fields: map[string]string{
						"foo": "bar",
					},
				},
			},
			verbose:  true,
			expected: fmt.Errorf(debugHeader+verboseClientError(newAPIError(http.StatusUnprocessableEntity))+": validation error: %v: payload error", map[string]string{"foo": "bar"}),
		},
	}

	for desc, test := range tests {
		res := HandleClientError(test.err, test.verbose)
		if fmt.Sprintf("%v", test.expected) != fmt.Sprintf("%v", res) {
			t.Errorf("%s: errors do not match: expected '%v' but got '%v'", desc, test.expected, res)
		}
	}
	return
}

func TestHandleError(t *testing.T) {
	err := errors.Wrap(errors.Wrap(errors.New("error1"), "error2"), "error3")
	h := HandleError(err, false)
	expected := "error3"
	if fmt.Sprintf("%v", h) != expected {
		t.Errorf("errors do not match: expected '%v' but got '%v'", expected, h)
	}
}
