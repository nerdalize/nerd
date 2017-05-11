package v1data

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/nerdalize/nerd/nerd/client"
)

const (
	//IndexObjectKey is the key of the object that contains an index of all the chunks of a dataset.
	IndexObjectKey = "index"
	//UploadPolynomal is the polynomal that is used for chunked uploading.
	UploadPolynomal = 0x3DA3358B4DC173
)

//Key is the identifier of a chunk of data.
type Key [sha256.Size]byte

//ToString returns the string representation of a key.
func (k Key) ToString() string {
	return fmt.Sprintf("%x", k)
}

func KeyFromString(key string) (k Key, err error) {
	bytes, err := hex.DecodeString(key)
	if err != nil {
		return ZeroKey, fmt.Errorf("could not decode key '%v'", key)
	}
	copy(k[:], bytes)
	return k, nil
}

//ZeroKey is an empty key.
var ZeroKey = Key{}

//KeyReader can be implemented by objects capable of reading Keys.
type KeyReader interface {
	ReadKey() (Key, error)
}

//KeyWriter can be implemented by objects capable of writing Keys.
type KeyWriter interface {
	WriteKey(Key) error
}

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
