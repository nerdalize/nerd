package transfer

import (
	"context"
	"errors"
	"io"
	"io/ioutil"

	"github.com/nerdalize/nerd/pkg/transfer/archiver"
	"github.com/nerdalize/nerd/pkg/transfer/store"
)

//Reporter handles progress reporting
type Reporter interface {
	transferarchiver.Reporter

	HandledKey(key string)
	StartUploadProgress(label string, total int64, rr io.Reader) io.Reader
	StopUploadProgress()
	StartDownloadProgress(label string, total int64) io.Writer
	StopDownloadProgress()
}

//NewDiscardReporter discards all progress reports
func NewDiscardReporter() *DiscardReporter {
	return &DiscardReporter{}
}

//DiscardReporter is a reporter that discards
type DiscardReporter struct{}

//HandledKey discards the information a key was handled
func (r *DiscardReporter) HandledKey(key string) {}

//StartArchivingProgress is called when archiving has started and total size is known
func (r *DiscardReporter) StartArchivingProgress(label string, total int64) io.Writer {
	return ioutil.Discard
}

//StartUploadProgress is called when upload has started while total size is known
func (r *DiscardReporter) StartUploadProgress(label string, total int64, rr io.Reader) io.Reader {
	return rr
}

//StopUploadProgress is called when uploading has stopped
func (r *DiscardReporter) StopUploadProgress() {}

//StopArchivingProgress is called when archiving has stoppped
func (r *DiscardReporter) StopArchivingProgress() {}

//StartDownloadProgress will start the download progress
func (r *DiscardReporter) StartDownloadProgress(label string, total int64) io.Writer {
	return ioutil.Discard
}

//StopDownloadProgress will stop the download progress
func (r *DiscardReporter) StopDownloadProgress() {}

func (r *DiscardReporter) StartUnarchivingProgress(label string, total int64, rr io.Reader) io.Reader {
	return rr
}

func (r *DiscardReporter) StopUnarchivingProgress() {}

//Store provides an object storage interface
type Store interface {
	Head(ctx context.Context, k string) (size int64, err error)
	Get(ctx context.Context, key string, w io.WriterAt) error
	Put(ctx context.Context, key string, r io.ReadSeeker) error
	Del(ctx context.Context, key string) error
}

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
	Archive(path string, rep transferarchiver.Reporter, fn func(k string, r io.ReadSeeker, nbytes int64) error) error
	Unarchive(path string, rep transferarchiver.Reporter, fn func(k string, w io.WriterAt) error) error
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
