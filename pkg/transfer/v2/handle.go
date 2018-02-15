package transfer

import (
	"context"
	"io"

	"github.com/pkg/errors"
)

//StdHandle provides a standard implementation for handling datasets
type StdHandle struct {
	Meta
	store    Store
	archiver Archiver
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

	//set size to 0
	err = h.UpdateMeta(ctx, 0)
	if err != nil {
		return errors.Wrapf(err, "failed to update metadata")
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

	err = h.UpdateMeta(ctx, wc.total)
	if err != nil {
		return errors.Wrapf(err, "failed to update metadata")
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

	return nil
}
