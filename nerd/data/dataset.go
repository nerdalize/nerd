package data

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

const (
	//newline is the separator for Keys in the metadata file.
	newline  = '\n'
	newlineS = "\n"

	//MetadataObjectKey is the key of the S3 object that holds the metadata of a dataset.
	MetadataObjectKey = "metadata"
)

//Key is the identifier of a chunk of data.
type Key [sha256.Size]byte

//ToString returns the string representation of a key.
func (k Key) ToString() string {
	return fmt.Sprintf("%x", k)
}

//ZeroKey is an empty key.
var ZeroKey = Key{}

//KeyReadWriter is an interface that wraps the read and write methods for Keys.
//
//Read always returns a Key unless no more keys are available. In this case Read returns ZeroKey and an io.EOF error.
//Write writes a Key and returns an error if something goes wrong.
type KeyReadWriter interface {
	ReadKey() (Key, error)
	WriteKey(Key) error
}

//Metadata describes a dataset. It contains a header with different properties of the dataset and a
//KeyReadWriter which is used to keep track of the list of Keys (chunks) of the dataset.
type Metadata struct {
	Header *MetadataHeader
	KeyReadWriter
}

//MetadataHeader is the header of the dataset metadata.
type MetadataHeader struct {
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
	Size    int64     `json:"size"`
}

//NewMetadata returns a new Metadata.
func NewMetadata(header *MetadataHeader, krw KeyReadWriter) *Metadata {
	return &Metadata{
		Header:        header,
		KeyReadWriter: krw,
	}
}

//NewMetadataFromReader reads the header from a io.Reader and creates a streamingKeyReader to read the Keys from the reader.
func NewMetadataFromReader(r io.Reader) (*Metadata, error) {
	bufr := bufio.NewReader(r)
	headerData, err := bufr.ReadBytes(newline)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read dataset header")
	}
	header := new(MetadataHeader)
	err = json.Unmarshal(headerData, header)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal dataset header")
	}
	return &Metadata{
		Header:        header,
		KeyReadWriter: newStreamingKeyReader(bufr),
	}, nil
}

//ToString converts a Metadata to a string
func (m *Metadata) ToString() (string, error) {
	var s []string
	b, err := json.Marshal(m.Header)
	if err != nil {
		return "", errors.Wrap(err, "failed to convert header to JSON")
	}
	s = append(s, string(b))
	for {
		k, err := m.ReadKey()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", errors.Wrap(err, "failed to read key")
		}
		s = append(s, k.ToString())
	}
	return strings.Join(s, newlineS), nil
}

//streamingKeyReader reads keys from a buffered Reader.
type streamingKeyReader struct {
	*sync.Mutex
	r *bufio.Reader
}

//NewStreamingKeyReader creates a new streamingKeyReader.
func newStreamingKeyReader(r *bufio.Reader) *streamingKeyReader {
	return &streamingKeyReader{
		Mutex: new(sync.Mutex),
		r:     r,
	}
}

//ReadKey reads a key from the bufio.Reader.
//ReadKey returns an io.EOF error if no more Keys are available.
func (kw *streamingKeyReader) ReadKey() (Key, error) {
	kw.Lock()
	defer kw.Unlock()
	line, err := kw.r.ReadString(newline)
	line = strings.Replace(line, newlineS, "", 1)
	if err != nil {
		if err != io.EOF {
			return ZeroKey, errors.Wrap(err, "failed to read key from input stream")
		}
		if line == "" {
			return ZeroKey, io.EOF
		}
	}
	bytes, err := hex.DecodeString(line)
	if err != nil {
		return ZeroKey, errors.Wrapf(err, "could not decode key string '%v'", line)
	}
	var k Key
	copy(k[:], bytes)
	return k, nil
}

//WriteKey writing keys to a streamingKeyReader is not supported.
func (kw *streamingKeyReader) WriteKey(k Key) error {
	return errors.New("streamingKeyReader does not support writes")
}

//BufferedKeyReadWriter contains an internal buffer of Keys which could be read and written to.
type BufferedKeyReadWriter struct {
	*sync.Mutex
	pos int
	M   map[Key]struct{}
	L   []Key
}

//NewBufferedKeyReadWiter creates a new BufferedKeyReadWriter.
func NewBufferedKeyReadWiter() *BufferedKeyReadWriter {
	return &BufferedKeyReadWriter{Mutex: &sync.Mutex{}, M: map[Key]struct{}{}}
}

//WriteKey writes a Key to the buffer.
func (kw *BufferedKeyReadWriter) WriteKey(k Key) error {
	kw.Lock()
	defer kw.Unlock()
	if _, ok := kw.M[k]; ok {
		return nil
	}

	kw.M[k] = struct{}{}
	kw.L = append(kw.L, k)
	return nil
}

//ReadKey reads a Key from the buffer and returns an io.EOF if no more Keys are available.
func (kw *BufferedKeyReadWriter) ReadKey() (k Key, err error) {
	kw.Lock()
	defer kw.Unlock()
	if kw.pos == len(kw.L) {
		return ZeroKey, io.EOF
	}

	k = kw.L[kw.pos]
	kw.pos = kw.pos + 1
	return k, nil
}
