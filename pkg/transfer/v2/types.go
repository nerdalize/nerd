package transfer

import (
	"context"
	"errors"
	"io"
)

//Reporter handles progress reporting
type Reporter interface {
	HandledKey(key string)

	//@TODO report the objects that were accessed
	//@TODO think of what interface we would like here
}

//DiscardReporter discards all progress reports
func NewDiscardReporter() *DiscardReporter {
	return &DiscardReporter{}
}

type DiscardReporter struct{}

//HandledKey discards the information a key was handled
func (r *DiscardReporter) HandledKey(key string) {}

//A Handle provides interactions with a dataset
type Handle interface {
	io.Closer
	Name() string
	Clear(ctx context.Context, reporter Reporter) error
	Push(ctx context.Context, fromPath string, rep Reporter) error
	Pull(ctx context.Context, toPath string, rep Reporter) error
}

//Manager provides access to Transfer handles, this allows parallel
//access to datasets from multiple clients
type Manager interface {
	Create(ctx context.Context, name string, st StoreType, at ArchiverType, opts map[string]string) (Handle, error) //must not exist, name is unique, claims dataset handle
	Open(ctx context.Context, name string) (Handle, error)                                                          //must exist, claims dataset handle
	Remove(ctx context.Context, name string) error
	Info(ctx context.Context, name string) (size uint64, err error)
}

//Store provides an object storage interface
type Store interface {
	Get(ctx context.Context, key string, w io.WriterAt) error
	Put(ctx context.Context, key string, r io.Reader) error
	Del(ctx context.Context, key string) error
}

//Archiver allows archiving a directory
type Archiver interface {
	Index(fn func(k string) error) error
	Archive(path string, fn func(k string, r io.Reader) error) error
	Unarchive(path string, fn func(k string, w io.WriterAt) error) error
}

//StoreType determines what type the object store will be
type StoreType string

const (
	//StoreTypeS3 uses a AWS S3 store
	StoreTypeS3 StoreType = "s3"
)

//ArchiverType determines what type the object store will be
type ArchiverType string

const (
	//ArchiverTypeTar uses the tar archiving format
	ArchiverTypeTar ArchiverType = "tar"
)

//CreateStore will creates one of the standard storews with the provided options
func CreateStore(st StoreType, opts map[string]string) (Store, error) {
	switch st {
	case StoreTypeS3:
		return CreateS3Store(opts)
	default:
		return nil, errors.New("unsupported store")
	}
}

//CreateArchiver will creates one of the standard storews with the provided options
func CreateArchiver(at ArchiverType, opts map[string]string) (Archiver, error) {
	switch at {
	case ArchiverTypeTar:
		return CreateTarArchiver(opts)
	default:
		return nil, errors.New("unsupported archiver")
	}
}
