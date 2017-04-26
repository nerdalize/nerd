package v1data

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/nerdalize/nerd/nerd/client"
)

const (
	//MetadataObjectKey is the key of the S3 object that holds the metadata of a dataset.
	MetadataObjectKey = "metadata"
	IndexObjectKey    = "index"
)

//Key is the identifier of a chunk of data.
type Key [sha256.Size]byte

//ToString returns the string representation of a key.
func (k Key) ToString() string {
	return fmt.Sprintf("%x", k)
}

//ZeroKey is an empty key.
var ZeroKey = Key{}

type KeyReader interface {
	ReadKey() (Key, error)
}

type KeyWriter interface {
	WriteKey(Key) error
}

//Metadata describes a dataset. It contains a header with different properties of the dataset and a
//KeyReadWriter which is used to keep track of the list of Keys (chunks) of the dataset.
type Metadata struct {
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
	Size    int64     `json:"size"`
}

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

//
// //NewMetadata returns a new Metadata.
// func NewMetadata(header *MetadataHeader, krw KeyReadWriter) *Metadata {
// 	return &Metadata{
// 		Header:        header,
// 		KeyReadWriter: krw,
// 	}
// }
//
// //NewMetadataFromReader reads the header from a io.Reader and creates a streamingKeyReader to read the Keys from the reader.
// func NewMetadataFromReader(r io.Reader) (*Metadata, error) {
// 	bufr := bufio.NewReader(r)
// 	headerData, err := bufr.ReadBytes(newline)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to read dataset header")
// 	}
// 	header := new(MetadataHeader)
// 	err = json.Unmarshal(headerData, header)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "failed to unmarshal dataset header")
// 	}
// 	return &Metadata{
// 		Header:        header,
// 		KeyReadWriter: newStreamingKeyReader(bufr),
// 	}, nil
// }
//
// //ToString converts a Metadata to a string
// func (m *Metadata) ToString() (string, error) {
// 	var s []string
// 	b, err := json.Marshal(m.Header)
// 	if err != nil {
// 		return "", errors.Wrap(err, "failed to convert header to JSON")
// 	}
// 	s = append(s, string(b))
// 	for {
// 		k, err := m.ReadKey()
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			return "", errors.Wrap(err, "failed to read key")
// 		}
// 		s = append(s, k.ToString())
// 	}
// 	return strings.Join(s, newlineS), nil
// }
//
// //streamingKeyReader reads keys from a buffered Reader.
// type streamingKeyReader struct {
// 	*sync.Mutex
// 	r *bufio.Reader
// }
//
// //NewStreamingKeyReader creates a new streamingKeyReader.
// func newStreamingKeyReader(r *bufio.Reader) *streamingKeyReader {
// 	return &streamingKeyReader{
// 		Mutex: new(sync.Mutex),
// 		r:     r,
// 	}
// }
//
// //ReadKey reads a key from the bufio.Reader.
// //ReadKey returns an io.EOF error if no more Keys are available.
// func (kw *streamingKeyReader) ReadKey() (Key, error) {
// 	kw.Lock()
// 	defer kw.Unlock()
// 	line, err := kw.r.ReadString(newline)
// 	line = strings.Replace(line, newlineS, "", 1)
// 	if err != nil {
// 		if err != io.EOF {
// 			return ZeroKey, errors.Wrap(err, "failed to read key from input stream")
// 		}
// 		if line == "" {
// 			return ZeroKey, io.EOF
// 		}
// 	}
// 	bytes, err := hex.DecodeString(line)
// 	if err != nil {
// 		return ZeroKey, errors.Wrapf(err, "could not decode key string '%v'", line)
// 	}
// 	var k Key
// 	copy(k[:], bytes)
// 	return k, nil
// }
//
// //WriteKey writing keys to a streamingKeyReader is not supported.
// func (kw *streamingKeyReader) WriteKey(k Key) error {
// 	return errors.New("streamingKeyReader does not support writes")
// }
//
// //BufferedKeyReadWriter contains an internal buffer of Keys which could be read and written to.
// type BufferedKeyReadWriter struct {
// 	*sync.Mutex
// 	pos int
// 	M   map[Key]struct{}
// 	L   []Key
// }
//
// //NewBufferedKeyReadWiter creates a new BufferedKeyReadWriter.
// func NewBufferedKeyReadWiter() *BufferedKeyReadWriter {
// 	return &BufferedKeyReadWriter{Mutex: &sync.Mutex{}, M: map[Key]struct{}{}}
// }
//
// //WriteKey writes a Key to the buffer.
// func (kw *BufferedKeyReadWriter) WriteKey(k Key) error {
// 	kw.Lock()
// 	defer kw.Unlock()
// 	if _, ok := kw.M[k]; ok {
// 		return nil
// 	}
//
// 	kw.M[k] = struct{}{}
// 	kw.L = append(kw.L, k)
// 	return nil
// }
//
// //ReadKey reads a Key from the buffer and returns an io.EOF if no more Keys are available.
// func (kw *BufferedKeyReadWriter) ReadKey() (k Key, err error) {
// 	kw.Lock()
// 	defer kw.Unlock()
// 	if kw.pos == len(kw.L) {
// 		return ZeroKey, io.EOF
// 	}
//
// 	k = kw.L[kw.pos]
// 	kw.pos = kw.pos + 1
// 	return k, nil
// }
