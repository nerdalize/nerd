package v1datatransfer

import (
	"bytes"
	"context"
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
	progressCh        chan<- int64
}

//start starts the upload process
func (p *uploadProcess) start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	type countResult struct {
		total int64
		err   error
	}

	// pipeline: tar | count | chunk+upload | index
	tarCountPipe := newPipe()
	countChunksPipe := newPipe()
	chunksIndexPipe := newPipe()
	countReader := io.TeeReader(tarCountPipe.r, countChunksPipe.w)

	doneCh := make(chan error)
	countCh := make(chan countResult)

	// heartbeat
	go sendHeartbeat(ctx, cancel, doneCh, p.batchClient, p.dataset.ProjectID, p.dataset.DatasetID, p.heartbeatInterval)

	// chunks | index
	go func() {
		doneCh <- uploadIndex(ctx, p.dataClient, chunksIndexPipe.r, p.dataset.Bucket, p.dataset.DatasetRoot)
	}()
	// count | chunks
	go func() {
		defer chunksIndexPipe.w.Close()
		kw := v1data.NewIndexWriter(chunksIndexPipe.w)
		err := uploadChunks(ctx, p.dataClient, countChunksPipe.r, kw, p.dataset.Bucket, p.dataset.ProjectRoot, p.concurrency, p.progressCh)
		if err != nil {
			doneCh <- err
		}
	}()
	// tar | count
	go func() {
		total, err := countBytes(ctx, countReader)
		countCh <- countResult{
			err:   err,
			total: total,
		}
	}()
	// tar
	go func() {
		defer tarCountPipe.w.Close()
		defer countChunksPipe.w.Close()
		err := tardir(ctx, p.localDir, tarCountPipe.w)
		if err != nil {
			doneCh <- err
		}
	}()

	err := <-doneCh
	if err != nil {
		return err
	}

	cres := <-countCh
	if cres.err != nil {
		return errors.Wrap(err, "failed to calculate dataset size")
	}
	err = uploadMetadata(ctx, p.dataClient, cres.total, p.dataset.Bucket, p.dataset.DatasetRoot)
	if err != nil {
		return err
	}

	_, err = p.batchClient.SendUploadSuccess(p.dataset.ProjectID, p.dataset.DatasetID)
	if err != nil {
		return errors.Wrap(err, "failed to send dataset success message")
	}

	return nil
}

//uploadChunks uploads data from r in a chunked way
func uploadChunks(ctx context.Context, dataClient *v1data.Client, r io.Reader, kw v1data.KeyWriter, bucket, root string, concurrency int, progressCh chan<- int64) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
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
		key := path.Join(root, k.ToString())
		exists, err := dataClient.Exists(ctx, bucket, key) //check existence
		if err != nil {
			it.resCh <- &result{errors.Wrapf(err, "failed to check existence of '%x'", k), v1data.ZeroKey}
			return
		}

		if !exists {
			err = dataClient.Upload(ctx, bucket, key, bytes.NewReader(it.chunk)) //if not exists put
			if err != nil {
				it.resCh <- &result{errors.Wrapf(err, "failed to put chunk '%x'", k), v1data.ZeroKey}
				return
			}
		}
		if progressCh != nil {
			progressCh <- int64(len(it.chunk))
		}

		it.resCh <- &result{nil, k}
	}

	//fan out
	itemCh := make(chan *item, concurrency)
	go func() {
		defer close(itemCh)
		buf := make([]byte, chunker.MaxSize)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				chunk, err := chkr.Next(buf)
				if err != nil {
					if err != io.EOF {
						itemCh <- &item{err: err}
					}
					return
				}

				it := &item{
					chunk: make([]byte, chunk.Length),
					resCh: make(chan *result),
				}

				copy(it.chunk, chunk.Data) //underlying buffer is switched out

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

			err := kw.WriteKey(res.k)
			if err != nil {
				return errors.Wrapf(err, "failed to write key")
			}
		}
	}

	return nil
}

//uploadIndex uploads the index object with all the keys
func uploadIndex(ctx context.Context, dataClient *v1data.Client, r io.Reader, bucket, root string) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return errors.Wrap(err, "failed to read keys")
	}
	return dataClient.Upload(ctx, bucket, path.Join(root, v1data.IndexObjectKey), bytes.NewReader(b))
}

//uploadMetadata uploads the metadata object
func uploadMetadata(ctx context.Context, dataClient *v1data.Client, total int64, bucket, root string) error {
	metadata := &v1datapayload.Metadata{
		Size:    total,
		Created: time.Now(),
		Updated: time.Now(),
	}
	err := dataClient.MetadataUpload(ctx, bucket, root, metadata)
	if err != nil {
		return errors.Wrap(err, "failed to upload metadata")
	}
	return nil
}

//sendHeartbeat sends a heartbeat and sleeps for the given interval
func sendHeartbeat(ctx context.Context, cancel context.CancelFunc, doneCh chan error, batchClient *v1batch.Client, projectID, datasetID string, interval time.Duration) {
	ticker := time.Tick(interval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker:
			out, err := batchClient.SendUploadHeartbeat(projectID, datasetID)
			if err == nil && out.HasExpired {
				doneCh <- fmt.Errorf("upload failed because the server could not be reached for too long")
				cancel()
			}
		}
	}
}
