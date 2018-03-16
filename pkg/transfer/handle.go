package transfer

import (
	"context"
	"io"

	"github.com/pkg/errors"
)

//HandleDelegate allows customization of lifecycle events, these
//events can be handled inside the lock of the handle
type HandleDelegate interface {
	PostClean(ctx context.Context) error             //eg, set size to
	PostPush(ctx context.Context, size uint64) error //eg, set new size
	PostPull(ctx context.Context) error
	PostClose() error //eg release the lock
}

//StdHandle provides a standard implementation for handling datasets
type StdHandle struct {
	name     string
	delegate HandleDelegate
	store    Store
	archiver Archiver
}

//CreateStdHandle sets up a standard implementation of the handle
func CreateStdHandle(name string, store Store, a Archiver, del HandleDelegate) (*StdHandle, error) {
	if store == nil || a == nil {
		return nil, errors.New("store")
	}

	return &StdHandle{name: name, store: store, archiver: a, delegate: del}, nil
}

//Name returns the name
func (h *StdHandle) Name() string { return h.name }

//Clear removes all objects related to a dataset
func (h *StdHandle) Clear(ctx context.Context, reporter Reporter) (err error) {

	if err = h.archiver.Index(func(k string) error {
		if err = h.store.Del(ctx, k); err != nil {
			return errors.Wrap(err, "failed to delete object key")
		}

		reporter.HandledKey(k)
		//@TODO inform reporter

		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to walk index")
	}

	if h.delegate != nil {
		if err = h.delegate.PostClean(ctx); err != nil {
			return errors.Wrap(err, "failed to run post clean delegate")
		}
	}

	return nil
}

type progressReader struct {
	io.ReadSeeker
	proxy io.Reader
}

func newProgressReader(w io.Writer, r io.ReadSeeker, proxy io.Reader) io.ReadSeeker {
	return &progressReader{ReadSeeker: r, proxy: io.TeeReader(proxy, w)}
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	return pr.proxy.Read(p)
}

//Push pushes new content from a local filesystem
func (h *StdHandle) Push(ctx context.Context, fromPath string, rep Reporter) (err error) {

	wc := &writeCounter{}
	if err = h.archiver.Archive(ctx, fromPath, rep, func(k string, r io.ReadSeeker, nbytes int64) error {

		//push bytes while counting the total number being pushed across all objects
		defer rep.StopUploadProgress()
		if err = h.store.Put(ctx, k, newProgressReader(wc, r, rep.StartUploadProgress(k, nbytes, r))); err != nil {
			return errors.Wrap(err, "failed to put object")
		}

		return nil
	}); err != nil {
		return errors.Wrapf(err, "failed to archive")
	}

	if h.delegate != nil {
		if err = h.delegate.PostPush(ctx, wc.total); err != nil {
			return errors.Wrap(err, "failed to run post push delegate")
		}
	}

	return nil
}

type progressWriter struct {
	io.WriterAt
	proxy io.Writer
}

func newProgressWriter(w io.WriterAt, proxy io.Writer) io.WriterAt {
	return &progressWriter{WriterAt: w, proxy: proxy}
}

func (pw *progressWriter) WriteAt(p []byte, off int64) (n int, err error) {
	pw.proxy.Write(p) //unconditionally also write to the progress proxy
	return pw.WriterAt.WriteAt(p, off)
}

//Pull content from the store to the local filesystem
func (h *StdHandle) Pull(ctx context.Context, toPath string, rep Reporter) (err error) {
	if err = h.archiver.Unarchive(ctx, toPath, rep, func(k string, w io.WriterAt) error {

		var total int64
		total, err = h.store.Head(ctx, k)
		if err != nil {
			return errors.Wrap(err, "failed to get object metadata")
		}

		pw := rep.StartDownloadProgress(k, total)
		defer rep.StopDownloadProgress()

		if err = h.store.Get(ctx, k, newProgressWriter(w, pw)); err != nil {
			return errors.Wrap(err, "failed to get object")
		}

		//@TODO update progress, per byte also while unarchiving

		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to unarchive")
	}

	if h.delegate != nil {
		if err = h.delegate.PostPull(ctx); err != nil {
			return errors.Wrap(err, "failed to run post pull delegate")
		}
	}

	return nil
}

//Close the handle performing any cleanup logic
func (h *StdHandle) Close() (err error) {
	if h.delegate != nil {
		if err = h.delegate.PostClose(); err != nil {
			return errors.Wrap(err, "failed to run post close delegate")
		}
	}

	return nil
}

//write counter discards every byte written but keeps a count
type writeCounter struct{ total uint64 }

func (wc *writeCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.total += uint64(n)
	return n, nil
}
