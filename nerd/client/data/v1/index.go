package v1data

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/nerdalize/nerd/nerd/client"
)

const (
	//IndexObjectKey is the key of the object that contains an index of all the chunks of a dataset.
	IndexObjectKey = "index"
)

//IndexReader can be used to read keys from the "index" object.
type IndexReader struct {
	s *bufio.Scanner
}

//NewIndexReader creates a new IndexReader.
func NewIndexReader(r io.Reader) *IndexReader {
	return &IndexReader{
		s: bufio.NewScanner(r),
	}
}

//ReadKey reads Keys from the provided io.Reader.
func (r *IndexReader) ReadKey() (Key, error) {
	if !r.s.Scan() {
		return ZeroKey, io.EOF
	}
	line := r.s.Text()
	bytes, err := hex.DecodeString(line)
	if err != nil {
		return ZeroKey, client.NewError(fmt.Sprintf("could not decode key string '%v'", line), err)
	}
	var k Key
	copy(k[:], bytes)
	return k, nil
}

//IndexWriter can be used to write keys to the "index" object.
type IndexWriter struct {
	w io.Writer
}

//NewIndexWriter creates a new IndexWriter.
func NewIndexWriter(w io.Writer) *IndexWriter {
	return &IndexWriter{
		w: w,
	}
}

//WriteKey writes a Key to the io.WriteCloser.
func (w *IndexWriter) WriteKey(k Key) error {
	_, err := w.w.Write([]byte(fmt.Sprintf("%v\n", k.ToString())))
	if err != nil {
		return client.NewError(fmt.Sprintf("failed to write key '%v' to writer", k.ToString()), err)
	}
	return nil
}
