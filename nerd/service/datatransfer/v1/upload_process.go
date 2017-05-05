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
	v1data "github.com/nerdalize/nerd/nerd/service/datatransfer/v1/client"
	v1datapayload "github.com/nerdalize/nerd/nerd/service/datatransfer/v1/client/payload"
	"github.com/pkg/errors"
	"github.com/restic/chunker"
)

type uploadProcess struct {
	batchClient       *v1batch.Client
	dataClient        *v1data.Client
	dataset           v1payload.DatasetSummary
	heartbeatInterval time.Duration
	localDir          string
	concurrency       int
	progressCh        chan int64
}

func (p *uploadProcess) sendHeartbeats(w *io.PipeWriter) {
	return
	sendHearbeats := true
	for sendHearbeats {
		time.Sleep(p.heartbeatInterval)
		out, err := p.batchClient.SendUploadHeartbeat(p.dataset.ProjectID, p.dataset.DatasetID)
		if err != nil {
			continue
		}
		if out.HasExpired {
			sendHearbeats = false
			w.CloseWithError(newPipeErr(fmt.Errorf("upload failed because the server could not be reached for too long")))
		}
	}
}

func (p *uploadProcess) uploadIndex(r io.Reader, doneCh chan error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		if isPipeErr(err) {
			doneCh <- err
			return
		}
		doneCh <- errors.Wrap(err, "failed to read keys")
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
		if p.progressCh != nil {
			p.progressCh <- int64(len(it.chunk))
		}

		it.resCh <- &result{nil, k}
	}

	//fan out
	itemCh := make(chan *item, p.concurrency)
	go func() {
		defer close(itemCh)
		buf := make([]byte, chunker.MaxSize)
		for {
			fmt.Println("Chunk before")
			chunk, err := chkr.Next(buf)
			fmt.Println("Chunk read")
			if err != nil {
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
			if isPipeErr(it.err) {
				w.CloseWithError(it.err)
				return
			}
			w.CloseWithError(newPipeErr(errors.Wrapf(it.err, "failed to iterate")))
			return
		}

		res := <-it.resCh
		if res.err != nil {
			w.CloseWithError(newPipeErr(res.err))
			return
		}

		err := kw.WriteKey(res.k)
		if err != nil {
			w.CloseWithError(newPipeErr(errors.Wrapf(err, "failed to write key")))
			return
		}
	}
}

func (p *uploadProcess) uploadMetadata(total int64) error {
	metadata := &v1datapayload.Metadata{
		Size:    total,
		Created: time.Now(),
		Updated: time.Now(),
	}
	err := p.dataClient.MetadataUpload(p.dataset.Bucket, p.dataset.DatasetRoot, metadata)
	if err != nil {
		// TODO: wrap error
		return err
	}
	return nil
}

func (p *uploadProcess) start() error {
	if p.progressCh != nil {
		defer close(p.progressCh)
	}
	// pipeline: tar -> count -> chunk+upload -> index
	tar_count := newPipe()
	count_chunks := newPipe()
	chunks_index := newPipe()
	countReader := io.TeeReader(tar_count.r, count_chunks.w)

	doneCh := make(chan error)
	countCh := make(chan countResult)

	// chunks -> index
	go p.uploadIndex(chunks_index.r, doneCh)
	// count -> chunks
	go p.uploadChunks(count_chunks.r, chunks_index.w)
	// tar -> count
	go countBytes(countReader, countCh)
	go p.sendHeartbeats(tar_count.w)
	err := tardir(p.localDir, tar_count.w)

	if err != nil && err != io.ErrClosedPipe {
		return errors.Wrapf(err, "failed to tar '%s'", p.localDir)
	}

	err = <-doneCh
	if err != nil {
		return err
	}

	cres := <-countCh
	if cres.err != nil {
		// TODO: wrap
		return err
	}

	err = p.uploadMetadata(cres.total)
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

type countResult struct {
	total int64
	err   error
}

//countBytes counts all bytes from a reader.
func countBytes(r io.Reader, resultCh chan countResult) {
	var total int64
	buf := make([]byte, 512*1024)
	for {
		n, err := io.ReadFull(r, buf)
		if err == io.ErrUnexpectedEOF {
			err = nil
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			resultCh <- countResult{
				total: 0,
				err:   errors.Wrap(err, "failed to read part of tar"),
			}
			return
		}
		total = total + int64(n)
	}

	resultCh <- countResult{
		total: total,
		err:   nil,
	}
	return
}
