package v1datatransfer

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"time"

	v1batch "github.com/nerdalize/nerd/nerd/client/batch/v1"
	v1payload "github.com/nerdalize/nerd/nerd/client/batch/v1/payload"
	v1data "github.com/nerdalize/nerd/nerd/client/data/v1"
	v1datapayload "github.com/nerdalize/nerd/nerd/client/data/v1/payload"
	"github.com/pkg/errors"
	"github.com/restic/chunker"
)

var heartbeatExpiredErr = fmt.Errorf("Upload failed because the server could not be reached for too long")

type pipeErr struct {
	error
}

func newPipeErr(err error) *pipeErr {
	return &pipeErr{err}
}

type uploadProcess struct {
	batchClient       *v1batch.Client
	dataClient        *v1data.Client
	dataset           v1payload.DatasetSummary
	heartbeatInterval time.Duration
	localDir          string
	concurrency       int
	progressCh        chan int64
}

func newUploadProcess(ds v1payload.DatasetSummary, concurrency int, progressCh chan int64) *uploadProcess {
	process := &uploadProcess{
		dataset:           ds,
		heartbeatInterval: 15 * time.Second,
		concurrency:       concurrency,
		progressCh:        progressCh,
	}
	return process
}

type pipe struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func newPipe() *pipe {
	pr, pw := io.Pipe()
	return &pipe{
		r: pr,
		w: pw,
	}
}

func (p *uploadProcess) sendHeartbeats(w *io.PipeWriter) {
	sendHearbeats := true
	for sendHearbeats {
		time.Sleep(p.heartbeatInterval)
		out, err := p.batchClient.SendUploadHeartbeat(p.dataset.ProjectID, p.dataset.DatasetID)
		if err != nil {
			continue
		}
		if out.HasExpired {
			sendHearbeats = false
			w.CloseWithError(heartbeatExpiredErr)
		}
	}
}

func (p *uploadProcess) uploadIndex(r io.Reader, doneCh chan error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		if _, ok := err.(*pipeErr); ok {
			doneCh <- err
		} else {
			doneCh <- errors.Wrap(err, "failed to read keys")
		}
		return
	}
	doneCh <- p.dataClient.Upload(p.dataset.Bucket, path.Join(p.dataset.DatasetRoot, v1data.IndexObjectKey), bytes.NewReader(b))
}

func (p *uploadProcess) uploadChunks(r io.Reader, w *io.PipeWriter) {
	kw := v1data.NewIndexWriter(w)
	chkr := chunker.New(r, chunker.Pol(v1data.UploadPolynomal))
	type result struct {
		err error
		k   v1data.Key
	}

	type item struct {
		chunk []byte
		size  int64
		resCh chan *result
		err   error
	}

	work := func(it *item) {
		k := v1data.Key(sha256.Sum256(it.chunk)) //hash
		key := path.Join(p.dataset.DatasetRoot, k.ToString())
		exists, err := p.dataClient.Exists(p.dataset.Bucket, key) //check existence
		if err != nil {
			it.resCh <- &result{errors.Wrapf(err, "failed to check existence of '%x'", k), v1data.ZeroKey}
			return
		}

		if !exists {
			err = p.dataClient.Upload(p.dataset.Bucket, key, bytes.NewReader(it.chunk)) //if not exists put
			if err != nil {
				it.resCh <- &result{errors.Wrapf(err, "failed to put chunk '%x'", k), v1data.ZeroKey}
				return
			}
		}
		p.progressCh <- int64(len(it.chunk))

		it.resCh <- &result{nil, k}
	}

	//fan out
	itemCh := make(chan *item, p.concurrency)
	go func() {
		defer close(itemCh)
		buf := make([]byte, chunker.MaxSize)
		for {
			chunk, err := chkr.Next(buf)
			if err != nil {
				if err == heartbeatExpiredErr {
					w.CloseWithError(newPipeErr(heartbeatExpiredErr))
					// TODO: stop p.uploadChunks
				}
				if err != io.EOF {
					itemCh <- &item{err: err}
				}
				break
			}

			it := &item{
				chunk: make([]byte, chunk.Length),
				resCh: make(chan *result),
			}

			copy(it.chunk, chunk.Data) //underlying buffer is switched out

			go work(it)  //create work
			itemCh <- it //send to fan-in thread for syncing results
		}
	}()

	//fan-in
	for it := range itemCh {
		if it.err != nil {
			// TODO: close pipewriter
			// return errors.Wrapf(it.err, "failed to iterate")
		}

		res := <-it.resCh
		if res.err != nil {
			// TODO: close pipewriter
			// return res.err
		}

		err := kw.WriteKey(res.k)
		if err != nil {
			// TODO: close pipewriter
			// return errors.Wrapf(err, "failed to write key")
		}
	}
}

func (p *uploadProcess) uploadMetadata() error {
	size, err := totalTarSize(p.localDir)
	if err != nil {
		//TODO: wrap error
		return err
	}
	metadata := &v1datapayload.Metadata{
		Size:    size,
		Created: time.Now(),
		Updated: time.Now(),
	}
	err = p.dataClient.MetadataUpload(p.dataset.Bucket, p.dataset.DatasetRoot, metadata)
	if err != nil {
		// TODO: wrap error
		return err
	}
	return nil
}

func (p *uploadProcess) upload() error {
	tar_chunks := newPipe()
	chunks_index := newPipe()

	doneCh := make(chan error)

	go p.uploadIndex(chunks_index.r, doneCh)
	go p.uploadChunks(tar_chunks.r, chunks_index.w)
	go p.sendHeartbeats(tar_chunks.w)
	err := tardir(p.localDir, tar_chunks.w)

	if err != nil && errors.Cause(err) != io.ErrClosedPipe {
		return errors.Wrapf(err, "failed to tar '%s'", p.localDir)
	}

	err = <-doneCh
	if err != nil {
		return err
	}
	return nil
}

func (p *uploadProcess) start() error {
	err := p.upload()
	if err != nil {
		return err
	}

	err = p.uploadMetadata()
	if err != nil {
		return err
	}

	_, err = p.batchClient.SendUploadSuccess(p.dataset.ProjectID, p.dataset.DatasetID)
	if err != nil {
		// TODO: wrap error
		return err
	}

	return nil
}
