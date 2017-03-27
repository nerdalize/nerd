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

type Key [sha256.Size]byte

func (k Key) ToString() string {
	return fmt.Sprintf("%x", k)
}

var ZeroKey = Key{}

type KeyReadWriter interface {
	ReadKey() (Key, error)
	WriteKey(Key) error
}

type DatasetHeader struct {
	Created time.Time `json:"created_at"`
	Updated time.Time `json:"updated_at"`
	Size    int64     `json:"size"`
}

type Metadata struct {
	Header *DatasetHeader
	KeyReadWriter
}

func NewMetadata(header *DatasetHeader, krw KeyReadWriter) *Metadata {
	return &Metadata{
		Header:        header,
		KeyReadWriter: krw,
	}
}

func NewMetadataFromReader(r io.Reader) (*Metadata, error) {
	bufr := bufio.NewReader(r)
	// TODO: \n magic char
	headerData, err := bufr.ReadBytes('\n')
	if err != nil {
		return nil, errors.Wrap(err, "failed to read dataset header")
	}
	header := new(DatasetHeader)
	err = json.Unmarshal(headerData, header)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal dataset header")
	}
	return &Metadata{
		Header:        header,
		KeyReadWriter: NewStreamingKeyReader(bufr),
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
	return strings.Join(s, "\n"), nil
}

type streamingKeyReader struct {
	*sync.Mutex
	r *bufio.Reader
}

// TODO: Abstract root
func NewStreamingKeyReader(r *bufio.Reader) *streamingKeyReader {
	return &streamingKeyReader{
		Mutex: new(sync.Mutex),
		r:     r,
	}
}

func (kw *streamingKeyReader) ReadKey() (Key, error) {
	kw.Lock()
	defer kw.Unlock()
	line, err := kw.r.ReadString('\n')
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
func (kw *streamingKeyReader) WriteKey(k Key) error {
	return errors.New("streamingKeyReader does not support writes")
}

type bufferedKeyReadWriter struct {
	*sync.Mutex
	pos int
	M   map[Key]struct{}
	L   []Key
}

func NewBufferedKeyReadWiter() *bufferedKeyReadWriter {
	return &bufferedKeyReadWriter{Mutex: &sync.Mutex{}, M: map[Key]struct{}{}}
}

func (kw *bufferedKeyReadWriter) WriteKey(k Key) error {
	kw.Lock()
	defer kw.Unlock()
	if _, ok := kw.M[k]; ok {
		return nil
	}

	kw.M[k] = struct{}{}
	kw.L = append(kw.L, k)
	return nil
}

func (kw *bufferedKeyReadWriter) ReadKey() (k Key, err error) {
	kw.Lock()
	defer kw.Unlock()
	if kw.pos == len(kw.L) {
		return ZeroKey, io.EOF
	}

	k = kw.L[kw.pos]
	kw.pos = kw.pos + 1
	return k, nil
}
