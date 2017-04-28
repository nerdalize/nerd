package v1data

import (
	"bytes"
	"io"
	"io/ioutil"
	"sync"
	"testing"
)

const KiB = 1024
const MiB = KiB * 1024

type fakeDataOps struct {
	data map[string][]byte
	m    *sync.Mutex
}

func (f *fakeDataOps) Upload(bucket, key string, body io.ReadSeeker) error {
	f.m.Lock()
	defer f.m.Unlock()

	dat, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	f.data[bucket+"/"+key] = dat
	return nil
}
func (f *fakeDataOps) Download(bucket, key string) (body io.ReadCloser, err error) {
	f.m.Lock()
	defer f.m.Unlock()
	return ioutil.NopCloser(bytes.NewReader(f.data[bucket+"/"+key])), nil
}
func (f *fakeDataOps) Exists(bucket, key string) (exists bool, err error) {
	_, ok := f.data[bucket+"/"+key]
	return ok, nil
}

func newClient() *Client {
	return NewClient(&fakeDataOps{
		data: make(map[string][]byte),
		m:    &sync.Mutex{},
	})
}

type fakeChunker struct {
	chunks      [][]byte
	chunkLength uint
	current     uint
	chunksSize  uint
	m           *sync.Mutex
}

func (f *fakeChunker) Next() (data []byte, length uint, err error) {
	f.m.Lock()
	defer f.m.Unlock()
	if f.chunksSize == f.current {
		return []byte{}, 0, io.EOF
	}
	dat := f.chunks[f.current]
	f.current = f.current + 1
	return dat, f.chunkLength, nil
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

//TestUpDown tests up/download end-to-end with a mocked S3 interface.
func TestUpDown(t *testing.T) {
	chunks := [][]byte{
		[]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		[]byte{5, 1, 5, 3, 4, 5, 6, 7, 8, 9},
		[]byte{0, 1, 2, 4, 4, 5, 6, 7, 8, 9},
		[]byte{9, 1, 2, 3, 4, 5, 6, 6, 8, 9},
	}
	var expected []byte
	for _, chunk := range chunks {
		expected = append(expected, chunk...)
	}
	chunker := &fakeChunker{
		chunks:      chunks,
		chunkLength: 10,
		chunksSize:  4,
		current:     0,
		m:           &sync.Mutex{},
	}
	cl := newClient()
	keyrw := NewBufferedKeyReadWiter()
	progressCh := make(chan int64)
	go func() {
		for _ = range progressCh {
		}
	}()

	t.Run("upload", func(t *testing.T) {
		cl.ChunkedUpload(chunker, keyrw, 10, "bucket", "root", progressCh)
	})

	var output = bytes.NewBuffer([]byte{})
	t.Run("download", func(t *testing.T) {
		cl.ChunkedDownload(keyrw, output, 10, "bucket", "root", progressCh)
	})

	if !bytes.Equal(output.Bytes(), expected) {
		t.Error("downloaded data should be equal to expected")
	}
}
