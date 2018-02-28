package transferarchiver

import (
	"archive/tar"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	slashpath "path"

	"github.com/pkg/errors"
)

var (
	//TarArchiverKey configures the one object key returned by the tar Archiver
	TarArchiverKey = "archive.tar"

	//TarArchiverPathSeparator standardizes the header path for cross platform (un)archiving
	TarArchiverPathSeparator = "/"
)

//TarArchiver will archive a directory into a single tar file
type TarArchiver struct {
	keyPrefix string
}

//NewTarArchiver will setup the tar archiver
func NewTarArchiver(opts ArchiverOptions) (a *TarArchiver, err error) {
	a = &TarArchiver{keyPrefix: opts.TarArchiverKeyPrefix}

	if a.keyPrefix != "" && !strings.HasSuffix(a.keyPrefix, "/") {
		return nil, errors.Errorf("archiver key prefix must end with a forward slash")
	}

	return a, nil
}

//tempFile will setup a temproary file that can easily be cleaned
func (a *TarArchiver) tempFile() (f *os.File, clean func(), err error) {
	f, err = ioutil.TempFile("", "tar_archiver_")
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create temporary file")
	}

	return f, func() {
		_ = f.Close()
		_ = os.Remove(f.Name())
	}, nil
}

func (a *TarArchiver) checkTargetDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Wrap(err, "failed to open directory")
		}

		err = os.Mkdir(path, 0777) //@TODO decide on permissions before umask
		if err != nil {
			return errors.Wrap(err, "failed to create directory")
		}

		dir, err = os.Open(path)
		if err != nil {
			return errors.Wrap(err, "failed open created directory")
		}
	}

	fis, err := dir.Readdirnames(1)
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "failed to read directory")
	}

	if len(fis) > 0 {
		return errors.New("directory is not empty")
	}

	return nil
}

//Index calls 'fn' for all object keys that are part of the archive
func (a *TarArchiver) Index(fn func(k string) error) error {
	return fn(slashpath.Join(a.keyPrefix, TarArchiverKey))
}

//@TODO do we want to expose this through the interface?
func (a *TarArchiver) indexFS(path string, fn func(p string, fi os.FileInfo, err error) error) error {
	if err := filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if fi == nil || path == p {
			return nil //this is triggered when a directory doesn't have an executable bit
		}
		if err != nil {
			return err
		}

		return fn(p, fi, nil)
	}); err != nil {
		return err
	}

	return nil
}

//Archive will archive a directory at 'path' into readable objects 'r' and calls 'fn' for each
func (a *TarArchiver) Archive(path string, rep Reporter, fn func(k string, r io.ReadSeeker, nbytes int64) error) (err error) {
	var totalToTar int64
	if err = a.indexFS(path, func(p string, fi os.FileInfo, err error) error {
		if !fi.Mode().IsRegular() {
			return nil //nothing to write for dirs or symlinks
		}
		totalToTar += fi.Size()
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to index filesystem")
	}

	if totalToTar == 0 {
		return errors.New("cannot archive empty directory")
	}

	tmpf, clean, err := a.tempFile()
	if err != nil {
		return err
	}

	inc := rep.StartArchivingProgress(tmpf.Name(), totalToTar)

	defer clean()
	tw := tar.NewWriter(tmpf)
	defer tw.Close()

	if err = a.indexFS(path, func(p string, fi os.FileInfo, err error) error {
		rel, err := filepath.Rel(path, p)
		if err != nil {
			return errors.Wrap(err, "failed to determine relative path")
		}

		//write header with a filename that standardizes the Separator
		path := strings.Split(rel, string(filepath.Separator))
		hdr, err := tar.FileInfoHeader(fi, "") //@TODO find out how we handle symlinks
		if err != nil {
			return errors.Wrap(err, "failed to convert file info to tar header")
		}

		hdr.Name = strings.Join(path, TarArchiverPathSeparator)
		if err = tw.WriteHeader(hdr); err != nil {
			return errors.Wrap(err, "failed to write tar header")
		}

		if !fi.Mode().IsRegular() {
			return nil //nothing to write for dirs or symlinks
		}

		// open files for taring
		f, err := os.Open(p)
		defer f.Close()
		if err != nil {
			return errors.Wrap(err, "failed to open file for archiving")
		}

		// copy file data into tar writer
		var n int64
		if n, err = io.Copy(tw, f); err != nil {
			return errors.Wrap(err, "failed to copy file content to archive")
		}
		inc(n)
		return nil
	}); err != nil {
		return errors.Wrap(err, "failed to perform filesystem walk")
	}
	err = tw.Flush()
	if err != nil {
		return errors.Wrap(err, "failed to flush tar writer to disk")
	}

	_, err = tmpf.Seek(0, 0)
	if err != nil {
		return errors.Wrap(err, "failed to seek to beginning of file")
	}

	//stop progress reporting, we're done
	rep.StopArchivingProgress()
	fi, err := tmpf.Stat()
	if err != nil {
		return errors.Wrap(err, "failed to stat the temporary file")
	}

	return fn(slashpath.Join(a.keyPrefix, TarArchiverKey), tmpf, fi.Size())
}

//Unarchive will take a file system path and call 'fn' for each object that it needs for unarchiving.
//It writes to a temporary directory first and then moves this to the final location
func (a *TarArchiver) Unarchive(path string, rep Reporter, fn func(k string, w io.WriterAt) error) error {
	tmpf, clean, err := a.tempFile()
	if err != nil {
		return err
	}

	defer clean()
	err = fn(slashpath.Join(a.keyPrefix, TarArchiverKey), tmpf)
	if err != nil {
		return errors.Wrap(err, "failed to download to temporary file")
	}

	_, err = tmpf.Seek(0, 0)
	if err != nil {
		return errors.Wrap(err, "failed to seek to the beginning of file")
	}

	fi, err := tmpf.Stat()
	if err != nil {
		return errors.Wrap(err, "failed to stat temporary file")
	}

	err = a.checkTargetDir(path)
	if err != nil {
		return err
	}

	pr := rep.StartUnarchivingProgress(tmpf.Name(), fi.Size(), tmpf)
	defer rep.StopUnarchivingProgress()

	tr := tar.NewReader(pr)
	for {
		hdr, err := tr.Next()
		switch {
		case err == io.EOF:
			return nil //EOF we're done here
		case err != nil:
			return errors.Wrap(err, "failed to read next header")
		case hdr == nil:
			continue
		}

		// the target location where the dir/file should be created
		parts := []string{path}
		parts = append(parts, strings.Split(hdr.Name, TarArchiverPathSeparator)...)
		target := filepath.Join(parts...)

		switch hdr.Typeflag {
		case tar.TypeDir: //if its a dir and it doesn't exist create it, no-op if it exists already
			err = os.MkdirAll(target, hdr.FileInfo().Mode())
			if err != nil {
				return errors.Wrap(err, "failed to create directory for entry found in tar file")
			}

		case tar.TypeReg: //regular file is written, must not exist yet
			if err = func() (err error) {
				f, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, hdr.FileInfo().Mode())
				if err != nil {
					return errors.Wrap(err, "failed to open new file for tar entry ")
				}

				defer f.Close()
				if _, err := io.Copy(f, tr); err != nil {
					return errors.Wrap(err, "failed to copy archived file content")
				}

				return nil
			}(); err != nil {
				return errors.Wrap(err, "failed to extract file")
			}
		}
	}
}
