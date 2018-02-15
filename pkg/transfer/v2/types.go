package transfer

import (
	"context"
	"io"
)

//Reporter handles progress reporting
type Reporter interface {
	//@TODO think of what interface we would like here
}

//DiscardReporter discards any progress reporting
func DiscardReporter() Reporter {
	return struct{}{}
}

//A Handle provides interactions with a dataset
type Handle interface {
	Meta
	Clear(ctx context.Context, reporter Reporter) error
	Push(ctx context.Context, fromPath string, rep Reporter) error
	Pull(ctx context.Context, toPath string, rep Reporter) error
}

//Meta interface provides metadat reading and updating
type Meta interface {
	io.Closer
	Name() string

	//@TODO the signature and method for metadata updates of an handle are experimental
	UpdateMeta(ctx context.Context, size uint64) error
	ReadMeta(ctx context.Context) (size uint64, err error)
}

//Manager provides access to Transfer handles, this allows parallel
//access to datasets from multiple clients
type Manager interface {
	Create(ctx context.Context, name string, st StoreType, at ArchiverType, opts map[string]string) (Handle, error) //must not exist, name is unique, claims dataset handle
	Open(ctx context.Context, name string) (Handle, error)                                                          //must exist, claims dataset handle
	Remove(ctx context.Context, name string) error                                                                  //must exist
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

//StoreFactory creates stores using an opaque set of options
type StoreFactory func(opts map[string]string) (Store, error)

//ArchiverFactory creates archivers based on opaque options map
type ArchiverFactory func(opts map[string]string) (Archiver, error)
