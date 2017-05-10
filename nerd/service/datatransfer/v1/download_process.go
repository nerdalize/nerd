package v1datatransfer

import (
	"context"
	"io"
	"io/ioutil"
	"path"

	v1batch "github.com/nerdalize/nerd/nerd/client/batch/v1"
	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
	v1data "github.com/nerdalize/nerd/nerd/service/datatransfer/v1/client"
	"github.com/pkg/errors"
)

type downloadProcess struct {
	batchClient v1batch.ClientInterface
	dataClient  *v1data.Client
	dataset     v1payload.DatasetSummary
	localDir    string
	concurrency int
	progressCh  chan<- int64
}

//start starts the download process
func (p *downloadProcess) start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	// pipeline: index | chunks | untar
	indexChunksPipe := newPipe()
	chunksUntarPipe := newPipe()

	doneCh := make(chan error)
	go func() {
		doneCh <- untardir(ctx, p.localDir, chunksUntarPipe.r)
	}()
	go func() {
		defer chunksUntarPipe.w.Close()
		kr := v1data.NewIndexReader(indexChunksPipe.r)
		err := downloadChunks(ctx, p.dataClient, kr, chunksUntarPipe.w, p.concurrency, p.dataset.Bucket, p.dataset.ProjectRoot, p.progressCh)
		if err != nil {
			doneCh <- err
		}
	}()
	go func() {
		defer indexChunksPipe.w.Close()
		err := downloadIndex(ctx, p.dataClient, indexChunksPipe.w, p.dataset.Bucket, p.dataset.DatasetRoot)
		if err != nil {
			doneCh <- err
		}
	}()

	err := <-doneCh
	if err != nil {
		return err
	}
	return nil
}

//downloadChunks downloads individual chunks and writes them to w
func downloadChunks(ctx context.Context, dataClient *v1data.Client, kr v1data.KeyReader, w io.Writer, concurrency int, bucket, root string, progressCh chan<- int64) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	type result struct {
		err   error
		chunk []byte
	}

	type item struct {
		k     v1data.Key
		resCh chan *result
		err   error
	}

	work := func(it *item) {
		r, err := dataClient.Download(ctx, bucket, path.Join(root, it.k.ToString()))
		if err != nil {
			it.resCh <- &result{errors.Wrapf(err, "failed to get key '%s'", it.k), nil}
			return
		}
		defer r.Close()

		chunk, err := ioutil.ReadAll(r)
		if err != nil {
			it.resCh <- &result{errors.Wrap(err, "failed to copy chunk to byte buffer"), nil}
		}

		if progressCh != nil {
			progressCh <- int64(len(chunk))
		}

		it.resCh <- &result{nil, chunk}
	}

	//fan out
	itemCh := make(chan *item, concurrency)
	go func() {
		defer close(itemCh)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				k, err := kr.ReadKey()
				if err != nil {
					if err != io.EOF {
						itemCh <- &item{err: err}
					}
					return
				}

				it := &item{
					k:     k,
					resCh: make(chan *result),
				}

				go work(it)  //create work
				itemCh <- it //send to fan-in thread for syncing results
			}
		}
	}()

	//fan-in
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case it := <-itemCh:
			if it == nil {
				return nil
			}
			if it.err != nil {
				return errors.Wrapf(it.err, "failed to iterate")
			}

			res := <-it.resCh
			if res.err != nil {
				return res.err
			}

			_, err := w.Write(res.chunk)
			if err != nil {
				return errors.Wrapf(err, "failed to write key")
			}
		}
	}

	return nil
}

//downloadIndex downloads the index object and writes it to w
func downloadIndex(ctx context.Context, dataClient *v1data.Client, w io.Writer, bucket, root string) error {
	body, err := dataClient.Download(ctx, bucket, path.Join(root, v1data.IndexObjectKey))
	if err != nil {
		return errors.Wrap(err, "failed download index")
	}
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return errors.Wrap(err, "failed to read body")
	}
	_, err = w.Write(b)
	if err != nil {
		return errors.Wrap(err, "failed to write body")
	}
	return nil
}
