package transfer

import (
	"context"
	"io"
)

//A Handle provides interactions with a dataset, it is protected by a
//distributed lock handed out by the manager
type Handle interface {
	Upload(fromPath string, progress chan<- struct{}) error
	Download(toPath string, progress chan<- struct{}) error
	io.Closer
}

//Manager provides access to Transfer handles, this allows parralell
//access to datasets from multiple clients
type Manager interface {
	Create(ctx context.Context, name string, st StoreType, at ArchiveType) (Handle, error) //must not exist, name is unique, claims dataset handle
	Open(ctx context.Context, name string) (Handle, error)                                 //must exist, claims dataset handle
	Remove(ctx context.Context, name string) error                                         //must exist
}

//Store provides an object storage interface
type Store interface {
	Get(key string, w io.WriterAt) error
	Put(key string, r io.Reader) error
	Del(key string) error
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
	ArchiverTypeTar StoreType = "tar"
)
