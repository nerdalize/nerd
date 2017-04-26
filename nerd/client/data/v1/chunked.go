package v1data

import (
	"bytes"
	"crypto/sha256"
	"io"
	"io/ioutil"
	"path"

	"github.com/pkg/errors"
)

const (
	//uploadPolynomal is the polynomal that is used for chunked uploading.
	UploadPolynomal = 0x3DA3358B4DC173
)

type Chunker interface {
	Next() (data []byte, length uint, err error)
}

//ChunkedUpload uploads data from a io.Reader (`r`) as a list of chunks. The Key of every chunk uploaded will be written to the KeyReadWriter (`kw`).
//ChunkedDownload reports its progress (the amount of bytes uploaded) to the progressCh.
//It will start a maximum of `concurrency` concurrent go routines to upload in paralllel.
//`root` is used as the root path of the chunk in S3. Root will be concatenated with the key to make the full S3 object path.
func (c *Client) ChunkedUpload(chunker Chunker, w KeyWriter, concurrency int, bucket, root string, progressCh chan<- int64) (err error) {
	// func (c *Client) ChunkedUpload(r io.Reader, kw data.KeyReadWriter, concurrency int, bucket, root string, progressCh chan<- int64) (err error) {
	type result struct {
		err error
		k   Key
	}

	type item struct {
		chunk []byte
		size  int64
		resCh chan *result
		err   error
	}

	work := func(it *item) {
		k := Key(sha256.Sum256(it.chunk)) //hash
		key := path.Join(root, k.ToString())
		exists, err := c.Exists(bucket, key) //check existence
		if err != nil {
			it.resCh <- &result{errors.Wrapf(err, "failed to check existence of '%x'", k), ZeroKey}
			return
		}

		if !exists {
			err = c.Upload(bucket, key, bytes.NewReader(it.chunk)) //if not exists put
			if err != nil {
				it.resCh <- &result{errors.Wrapf(err, "failed to put chunk '%x'", k), ZeroKey}
				return
			}
		}
		progressCh <- int64(len(it.chunk))

		it.resCh <- &result{nil, k}
	}

	//fan out
	itemCh := make(chan *item, concurrency)
	go func() {
		defer close(itemCh)
		for {
			data, len, err := chunker.Next()
			if err != nil {
				if err != io.EOF {
					itemCh <- &item{err: err}
				}
				break
			}

			it := &item{
				chunk: make([]byte, len),
				resCh: make(chan *result),
			}

			copy(it.chunk, data) //underlying buffer is switched out

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

		err = w.WriteKey(res.k)
		if err != nil {
			return errors.Wrapf(err, "failed to write key")
		}
	}

	return nil
}

//ChunkedDownload downloads a list of chunks and writes them to a io.Writer.
//ChunkedDownload reports its progress (the amount of bytes downloaded) to the progressCh.
//It will start a maximum of `concurrency` concurrent go routines to download in paralllel.
//`root` is used as the root path of the chunk in S3. Root will be concatenated with the Key read from `kr` to make the full S3 object path.
func (c *Client) ChunkedDownload(kr KeyReader, cw io.Writer, concurrency int, bucket, root string, progressCh chan<- int64) (err error) {
	type result struct {
		err   error
		chunk []byte
	}

	type item struct {
		k     Key
		resCh chan *result
		err   error
	}

	work := func(it *item) {
		r, err := c.Download(bucket, path.Join(root, it.k.ToString()))
		defer r.Close()
		if err != nil {
			it.resCh <- &result{errors.Wrapf(err, "failed to get key '%s'", it.k), nil}
			return
		}

		chunk, err := ioutil.ReadAll(r)
		if err != nil {
			it.resCh <- &result{errors.Wrap(err, "failed to copy chunk to byte buffer"), nil}
		}

		progressCh <- int64(len(chunk))

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

		_, err = cw.Write(res.chunk)
		if err != nil {
			return errors.Wrapf(err, "failed to write key")
		}
	}

	return nil
}
