package command

import (
	"fmt"
	"os"

	"github.com/nerdalize/nerd/nerd/payload"
)

type stdoutkw struct{}

func (kw *stdoutkw) Write(k string) (err error) {
	_, err = fmt.Fprintf(os.Stdout, "%v\n", k)
	return err
}

func HandleClientError(err error) error {
	if aerr, ok := err.(payload.APIError); ok {
		if len(aerr.APIError.Fields) > 0 && aerr.Response.StatusCode == 403 {
			return fmt.Errorf("Validation error: %v", aerr.APIError.Fields)
		}
	}
	return err
}
