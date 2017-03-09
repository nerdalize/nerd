package client

import (
	"crypto/ecdsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/dghubble/sling"
	"github.com/nerdalize/nerd/nerd/client/credentials"
	"github.com/nerdalize/nerd/nerd/payload"
)

//fakeProvider is a fake credentials provider for testing purposes.
type fakeProvider struct{}

//IsExpired stub
func (f *fakeProvider) IsExpired() bool {
	return true
}

//Retrieve returns empty token.
func (f *fakeProvider) Retrieve(pub *ecdsa.PublicKey) (*credentials.NerdAPIValue, error) {
	return &credentials.NerdAPIValue{
		NerdToken: "",
	}, nil
}

//newServer creates a new server that responds with the result on every request and sets a
//status code > 2xx when success is false.
func newServer(result interface{}, success bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if success {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		enc := json.NewEncoder(w)
		enc.Encode(result)
	}))
}

func TestDoRequest(t *testing.T) {
	testCases := map[string]struct {
		successResult      interface{}
		failureResult      *payload.Error
		successfullRequest bool
		errorMessage       string
	}{
		"success": {
			successResult:      "blaat",
			failureResult:      nil,
			successfullRequest: true,
			errorMessage:       "",
		},
		"payload error": {
			successResult: nil,
			failureResult: &payload.Error{
				Message: "error message",
				Fields: map[string]string{
					"field1": "cause1",
				},
			},
			successfullRequest: false,
			errorMessage:       "error message",
		},
	}
	for name, tc := range testCases {
		var svr *httptest.Server
		if tc.successfullRequest {
			svr = newServer(tc.successResult, tc.successfullRequest)
		} else {
			svr = newServer(tc.failureResult, tc.successfullRequest)
		}
		defer svr.Close()
		s := sling.New().Get(svr.URL)
		cl, err := NewNerdAPI(NerdAPIConfig{
			Credentials: credentials.NewNerdAPI(&ecdsa.PublicKey{}, &fakeProvider{}),
			URL:         svr.URL,
		})
		if err != nil {
			t.Fatalf("[%v]: Failed to create client", name)
		}
		err = cl.doRequest(s, tc.successResult)
		if err != nil {
			aerr, ok := err.(*APIError)
			if !ok {
				t.Fatalf("[%v]: could not cast error to *APIError", name)
				continue
			}
			if !strings.Contains(err.Error(), tc.errorMessage) {
				t.Errorf("[%v]: error message does not match, expected substring '%v' actual '%v'", name, tc.errorMessage, err.Error())
			}
			if perr, ok := aerr.Err.(*payload.Error); ok {
				if !reflect.DeepEqual(perr, tc.failureResult) {
					t.Errorf("[%v]: errors do not match, expected '%v' actual '%v'", name, tc.failureResult, perr)
				}
			}
		} else {
			if !tc.successfullRequest {
				t.Errorf("[%v]: request was supposed to fail but succeeded", name)
			}
		}
	}
}
