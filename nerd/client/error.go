package client

import (
	"fmt"
	"io"
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
