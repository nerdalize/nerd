package v1datatransfer

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

//tardir archives the given directory and writes bytes to w.
func tardir(ctx context.Context, dir string, w io.Writer) (err error) {
	tw := tar.NewWriter(w)
	err = filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if fi.Mode().IsDir() {
				return nil
			}

			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return errors.Wrapf(err, "failed to determine path '%s' relative to '%s'", path, dir)
			}

			f, err := os.Open(path)
			defer f.Close()
			if err != nil {
				return errors.Wrapf(err, "failed to open file '%s'", rel)
			}

			err = tw.WriteHeader(&tar.Header{
				Name:    rel,
				Mode:    int64(fi.Mode()),
				ModTime: fi.ModTime(),
				Size:    fi.Size(),
			})
			if err != nil {
				return errors.Wrapf(err, "failed to write tar header for '%s'", rel)
			}

			n, err := io.Copy(tw, f)

			if err != nil {
				return errors.Wrapf(err, "failed to write tar file for '%s'", rel)
			}

			if n != fi.Size() {
				return errors.Errorf("unexpected nr of bytes written to tar, saw '%d' on-disk but only wrote '%d', is directory '%s' in use?", fi.Size(), n, dir)
			}
			return nil
		}
	})

	if err != nil {
		return errors.Wrapf(err, "failed to walk dir '%s'", dir)
	}

	if err = tw.Close(); err != nil {
		return errors.Wrap(err, "failed to write remaining data")
	}

	return nil
}

//untardir untars an archive from the reader to a directory on disk.
func untardir(ctx context.Context, dir string, r io.Reader) (err error) {
	tr := tar.NewReader(r)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:

			hdr, err := tr.Next()
			if err != nil {
				if err == io.EOF {
					return nil
				}

				return errors.Wrap(err, "failed to read next tar header")
			}

			path := filepath.Join(dir, hdr.Name)
			err = os.MkdirAll(filepath.Dir(path), 0777)
			if err != nil {
				return errors.Wrap(err, "failed to create dirs")
			}

			f, err := safeFilePath(path, os.FileMode(hdr.Mode))
			if err != nil {
				return errors.Wrap(err, "failed to create tmp safe file")
			}

			defer f.Close()
			n, err := io.Copy(f, tr)
			if err != nil {
				return errors.Wrap(err, "failed to write file content to tmp file")
			}

			if n != hdr.Size {
				return errors.Errorf("unexpected nr of bytes written, wrote '%d' saw '%d' in tar hdr", n, hdr.Size)
			}

			err = os.Chtimes(path, time.Now(), hdr.ModTime)
			if err != nil {
				return errors.Wrap(err, "failed to change times of tmp file")
			}
		}
	}
}

//safeFilePath returns a unique filename for a given filepath.
//For example: file.txt will become file_(1).txt if file.txt is already present.
func safeFilePath(p string, perm os.FileMode) (*os.File, error) {
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err == nil {
		return f, nil
	}
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	filename := filepath.Base(p)
	ext := filepath.Ext(filename)
	clean := strings.TrimSuffix(filename, ext)
	re := regexp.MustCompile("_\\(\\d+\\)$")
	versionMatch := re.FindString(clean)
	version := 1
	if versionMatch != "" {
		oldVersion, _ := strconv.Atoi(strings.Trim(versionMatch, "_()"))
		clean = strings.TrimSuffix(clean, fmt.Sprintf("_(%v)", oldVersion))
		version = oldVersion + 1
	}
	newFilename := fmt.Sprintf("%s_(%v)%s", clean, version, ext)
	newPath := path.Join(filepath.Dir(p), newFilename)
	return safeFilePath(newPath, perm)
}

//countBytes counts all bytes from a reader.
func countBytes(ctx context.Context, r io.Reader) (total int64, err error) {
	done := false
	buf := make([]byte, 512*1024)
	for {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			n, err := io.ReadFull(r, buf)
			if err == io.ErrUnexpectedEOF {
				err = nil
				done = true
			}
			if err == io.EOF {
				err = nil
				done = true
			}
			if err != nil {
				err = errors.Wrap(err, "failed to read part of tar")
				done = true
			}
			total = total + int64(n)

			if done {
				return total, err
			}
		}
	}
}
