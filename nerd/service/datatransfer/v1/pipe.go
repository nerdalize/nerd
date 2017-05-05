package v1datatransfer

import "io"

type pipeErr struct {
	error
}

func newPipeErr(err error) *pipeErr {
	return &pipeErr{err}
}

func isPipeErr(err error) bool {
	_, ok := err.(*pipeErr)
	return ok
}

type pipe struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func newPipe() *pipe {
	pr, pw := io.Pipe()
	return &pipe{
		r: pr,
		w: pw,
	}
}
