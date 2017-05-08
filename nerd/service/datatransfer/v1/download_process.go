package v1datatransfer

import (
	"io"
	"io/ioutil"
	"path"

	v1batch "github.com/nerdalize/nerd/nerd/client/batch/v1"
	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
	v1data "github.com/nerdalize/nerd/nerd/service/datatransfer/v1/client"
	"github.com/pkg/errors"
)

type downloadProcess struct {
	batchClient *v1batch.Client
	dataClient  *v1data.Client
	dataset     v1payload.DatasetSummary
	localDir    string
	concurrency int
	progressCh  chan<- int64
}

//downloadChunks starts the download process of all chunks and closes w when it is done
func (p *downloadProcess) downloadChunks(r *io.PipeReader, w *io.PipeWriter) {
	kr := v1data.NewIndexReader(r)
	err := downloadChunks(p.dataClient, kr, w, p.concurrency, p.dataset.Bucket, p.dataset.ProjectRoot, p.progressCh)
	if err != nil {
		if isPipeErr(errors.Cause(err)) {
			w.CloseWithError(errors.Cause(err))
			return
		}
		w.CloseWithError(newPipeErr(err))
		return
	}
	w.Close()
}

//downloadIndex downloads the index object and closes w when it is done
func (p *downloadProcess) downloadIndex(w *io.PipeWriter) {
	err := downloadIndex(p.dataClient, w, p.dataset.Bucket, p.dataset.DatasetRoot)
	if err != nil {
		err = newPipeErr(errors.Wrap(err, "failed to download index file"))
		w.CloseWithError(err)
		return
	}
	w.Close()
	return
}

//start starts the download process
func (p *downloadProcess) start() error {
	// pipeline: index | chunks | untar
	indexChunksPipe := newPipe()
	chunksUntarPipe := newPipe()

	doneCh := make(chan error)
	go func() {
		doneCh <- untardir(p.localDir, chunksUntarPipe.r)
	}()
	go p.downloadChunks(indexChunksPipe.r, chunksUntarPipe.w)
	p.downloadIndex(indexChunksPipe.w)

	err := <-doneCh
	if err != nil {
		if isPipeErr(errors.Cause(err)) {
			return errors.Cause(err)
		}
		return err
	}
	return nil
}

//downloadChunks downloads individual chunks and writes them to w
func downloadChunks(dataClient *v1data.Client, kr v1data.KeyReader, w io.Writer, concurrency int, bucket, root string, progressCh chan<- int64) error {
	if progressCh != nil {
		defer close(progressCh)
	}
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
		r, err := dataClient.Download(bucket, path.Join(root, it.k.ToString()))
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
			k, err := kr.ReadKey()
			if err != nil {
				if err != io.EOF {
					itemCh <- &item{err: err}
				}
				break
			}

			it := &item{
				k:     k,
				resCh: make(chan *result),
			}

			go work(it)  //create work
			itemCh <- it //send to fan-in thread for syncing results
		}
	}()

	//fan-in
	for it := range itemCh {
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

	return nil
}

//downloadIndex downloads the index object and writes it to w
func downloadIndex(dataClient *v1data.Client, w io.Writer, bucket, root string) error {
	body, err := dataClient.Download(bucket, path.Join(root, v1data.IndexObjectKey))
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
