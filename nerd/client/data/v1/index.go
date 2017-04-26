package v1data

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/nerdalize/nerd/nerd/client"
)

const (
	//IndexObjectKey is the key of the S3 object that contains an index of all the chunks of a dataset.
	IndexObjectKey = "index"
)

type IndexReader struct {
	s *bufio.Scanner
}

func NewIndexReader(r io.Reader) *IndexReader {
	return &IndexReader{
		s: bufio.NewScanner(r),
	}
}

func (r *IndexReader) ReadKey() (Key, error) {
	if !r.s.Scan() {
		return ZeroKey, io.EOF
	}
	line := r.s.Text()
	bytes, err := hex.DecodeString(line)
	if err != nil {
		return ZeroKey, &client.Error{fmt.Sprintf("could not decode key string '%v'", line), err}
	}
	var k Key
	copy(k[:], bytes)
	return k, nil
}

type IndexWriter struct {
	w io.WriteCloser
}

func NewIndexWriter(w io.WriteCloser) *IndexWriter {
	return &IndexWriter{
		w: w,
	}
}

func (w *IndexWriter) WriteKey(k Key) error {
	_, err := w.w.Write([]byte(fmt.Sprintf("%v\n", k.ToString())))
	if err != nil {
		return &client.Error{fmt.Sprintf("failed to write key '%v' to writer", k.ToString()), err}
	}
	return nil
}

func (w *IndexWriter) Close() error {
	err := w.w.Close()
	if err != nil {
		return &client.Error{"failed to close index writer", err}
	}
	return nil
}
