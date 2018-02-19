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
	delegate HandleDelegate
	store    Store
	archiver Archiver
}

//CreateStdHandle sets up a standard implementation of the handle
func CreateStdHandle(store Store, a Archiver, del HandleDelegate) (*StdHandle, error) {
	if store == nil || a == nil {
		return nil, errors.New("store")
	}

	return &StdHandle{store: store, archiver: a, delegate: del}, nil
}

//Clear removes all objects related to a dataset
func (h *StdHandle) Clear(ctx context.Context, reporter Reporter) (err error) {

	if err = h.archiver.Index(func(k string) error {
		if err = h.store.Del(ctx, k); err != nil {
			return errors.Wrap(err, "failed to delete object key")
		}

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

//Push pushes new content from a local filesystem
func (h *StdHandle) Push(ctx context.Context, fromPath string, reporter Reporter) (err error) {

	wc := &writeCounter{}
	if err = h.archiver.Archive(fromPath, func(k string, r io.Reader) error {
		if err = h.store.Put(ctx, k, io.TeeReader(r, wc)); err != nil {
			return errors.Wrap(err, "failed to put object")
		}

		//@TODO update progress, make sure per byte?

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

//Pull content from the store to the local filesystem
func (h *StdHandle) Pull(ctx context.Context, toPath string, reporter Reporter) (err error) {
	if err = h.archiver.Unarchive(toPath, func(k string, w io.WriterAt) error {
		if err = h.store.Get(ctx, k, w); err != nil {
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
