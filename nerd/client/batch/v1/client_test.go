package v1batch

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
)

type logger struct{}

func (l *logger) Error(args ...interface{})            {}
func (l *logger) Debugf(a string, args ...interface{}) {}

type input struct {
	Field string `json:"field"`
}

type output struct {
	Field string `json:"field"`
}

type clErr struct {
	Msg string
}

type testCaseFields struct {
	jwt    string
	method string
	path   string
	input  interface{}
	output interface{}
}

func TestInterfaceImplementation(t *testing.T) {
	var v1 ClientInterface
	v1 = &Client{}
	_ = v1
}

func TestDataset(t *testing.T) {
	cases := map[string]struct {
		fields      *testCaseFields
		httpHandler func(*testing.T, *testCaseFields) http.Handler
		handler     func(*testing.T, *testCaseFields, error)
	}{
		"input": {
			fields: &testCaseFields{
				jwt:    "",
				method: http.MethodGet,
				path:   "/path",
				input:  &input{"InputField"},
				output: nil,
			},
			httpHandler: func(t *testing.T, f *testCaseFields) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					dec := json.NewDecoder(r.Body)
					in := &input{}
					err := dec.Decode(in)
					if err != nil {
						t.Errorf("Failed to decode body: %v", err)
						return
					}
					expected := f.input.(*input)
					if in.Field != expected.Field {
						t.Errorf("Expected InputField as input but got %v", in.Field)
					}
				})
			},
			handler: func(t *testing.T, f *testCaseFields, err error) {
				if err != nil {
					t.Errorf("unexpected error %v", err)
					return
				}
				if f.output != nil {
					t.Errorf("expected output to be nil but was %v", f.output)
				}
			},
		},
		"path": {
			fields: &testCaseFields{
				jwt:    "",
				method: http.MethodGet,
				path:   "path/path2",
				input:  nil,
				output: nil,
			},
			httpHandler: func(t *testing.T, f *testCaseFields) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path != "/"+f.path {
						t.Errorf("Expected path to be /path, but was %v", r.URL.Path)
					}
				})
			},
			handler: func(t *testing.T, f *testCaseFields, err error) {
				if err != nil {
					t.Errorf("unexpected error %v", err)
					return
				}
				if f.output != nil {
					t.Errorf("expected output to be nil but was %v", f.output)
				}
			},
		},
		"jwt": {
			fields: &testCaseFields{
				jwt:    "abc.def.ghi",
				method: http.MethodGet,
				path:   "",
				input:  nil,
				output: nil,
			},
			httpHandler: func(t *testing.T, f *testCaseFields) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Header.Get(AuthHeader) != "Bearer "+f.jwt {
						t.Errorf("Expected jwt 'bearer %v', but got '%v'", f.jwt, r.Header.Get(AuthHeader))
					}
				})
			},
			handler: func(t *testing.T, f *testCaseFields, err error) {
				if err != nil {
					t.Errorf("unexpected error %v", err)
					return
				}
				if f.output != nil {
					t.Errorf("expected output to be nil but was %v", f.output)
				}
			},
		},
		"output": {
			fields: &testCaseFields{
				jwt:    "",
				method: http.MethodGet,
				path:   "",
				input:  nil,
				output: &output{},
			},
			httpHandler: func(t *testing.T, f *testCaseFields) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					enc := json.NewEncoder(w)
					out := &output{"OutputField"}
					err := enc.Encode(out)
					if err != nil {
						t.Errorf("failed to encode output: %v", err)
					}
				})
			},
			handler: func(t *testing.T, f *testCaseFields, err error) {
				if err != nil {
					t.Errorf("unexpected error %v", err)
					return
				}
				out, _ := f.output.(*output)
				if out.Field != "OutputField" {
					t.Errorf("expected output to be OutputField but was %v", out.Field)
				}
			},
		},
		"method": {
			fields: &testCaseFields{
				jwt:    "",
				method: http.MethodPost,
				path:   "",
				input:  nil,
				output: nil,
			},
			httpHandler: func(t *testing.T, f *testCaseFields) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.Method != f.method {
						t.Errorf("expected method to be %v but was %v", f.method, r.Method)
					}
				})
			},
			handler: func(t *testing.T, f *testCaseFields, err error) {
				if err != nil {
					t.Errorf("unexpected error %v", err)
					return
				}
				if f.output != nil {
					t.Errorf("expected output to be nil but was %v", f.output)
				}
			},
		},
		"errResponse": {
			fields: &testCaseFields{
				jwt:    "",
				method: http.MethodGet,
				path:   "",
				input:  nil,
				output: nil,
			},
			httpHandler: func(t *testing.T, f *testCaseFields) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
					enc := json.NewEncoder(w)
					out := &v1payload.Error{
						Message: "error",
					}
					err := enc.Encode(out)
					if err != nil {
						t.Errorf("failed to encode output: %v", err)
					}
				})
			},
			handler: func(t *testing.T, f *testCaseFields, err error) {
				cerr, _ := err.(*HTTPError)
				if cerr.Err.Message != "error" {
					t.Errorf("expected error message 'error' but was %v", cerr.Err.Message)
				}
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ts := httptest.NewServer(tc.httpHandler(t, tc.fields))
			base, err := url.Parse(ts.URL)
			if err != nil {
				t.Fatalf("failed to parse url %v: %v", ts.URL, err)
			}
			cl := NewClient(ClientConfig{
				Base:        base,
				Logger:      &logger{},
				JWTProvider: NewStaticJWTProvider(tc.fields.jwt),
			})
			err = cl.doRequest(tc.fields.method, tc.fields.path, tc.fields.input, tc.fields.output)
			tc.handler(t, tc.fields, err)
		})
	}
}
