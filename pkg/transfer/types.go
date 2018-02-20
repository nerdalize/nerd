package transfer

import (
	"context"
	"errors"
	"io"

	"github.com/nerdalize/nerd/pkg/transfer/archiver"
	"github.com/nerdalize/nerd/pkg/transfer/store"
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

//Store provides an object storage interface
type Store interface {
	Get(ctx context.Context, key string, w io.WriterAt) error
	Put(ctx context.Context, key string, r io.Reader) error
	Del(ctx context.Context, key string) error
}

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
	Create(ctx context.Context, name string, sti transferstore.StoreOptions, sto transferarchiver.ArchiverOptions) (Handle, error) //must not exist, name is unique, claims dataset handle
	Open(ctx context.Context, name string) (Handle, error)                                                                         //must exist, claims dataset handle
	Remove(ctx context.Context, name string) error
	Info(ctx context.Context, name string) (size uint64, err error)
}

//Archiver allows archiving a directory
type Archiver interface {
	Index(fn func(k string) error) error
	Archive(path string, fn func(k string, r io.Reader) error) error
	Unarchive(path string, fn func(k string, w io.WriterAt) error) error
}

//CreateArchiver will creates one of the standard storews with the provided options
func CreateArchiver(opts transferarchiver.ArchiverOptions) (Archiver, error) {
	switch opts.Type {
	case transferarchiver.ArchiverTypeTar:
		return transferarchiver.NewTarArchiver(opts)
	default:
		return nil, errors.New("unsupported archiver")
	}
}

//CreateStore will creates of the standard stores based on the store options
func CreateStore(opts transferstore.StoreOptions) (Store, error) {
	switch opts.Type {
	case transferstore.StoreTypeS3:
		return transferstore.NewS3Store(opts)
	default:
		return nil, errors.New("unsupported store")
	}
}
