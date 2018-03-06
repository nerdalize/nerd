package transferarchiver_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	transfer "github.com/nerdalize/nerd/pkg/transfer"
	"github.com/nerdalize/nerd/pkg/transfer/archiver"
)

func archive(tb testing.TB, a transfer.Archiver, dir string, assertErr error) map[string][]byte {
	tb.Helper()
	rep := transfer.NewDiscardReporter()

	objs := map[string][]byte{}
	err := a.Archive(dir, rep, func(k string, r io.ReadSeeker, nbytes int64) error {
		buf := bytes.NewBuffer(nil)
		_, err := io.Copy(buf, r)

		objs[k] = buf.Bytes()
		return err
	})

	if !reflect.DeepEqual(err, assertErr) {
		tb.Fatalf("unexpected error, got: %v, want: %v", err, assertErr)
	}

	return objs
}

func TestTarArchiver(t *testing.T) {
	var (
		a   transfer.Archiver
		err error
	)

	rep := transfer.NewDiscardReporter()

	a, err = transferarchiver.NewTarArchiver(transferarchiver.ArchiverOptions{})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("archive non existing directory", func(t *testing.T) {
		archive(t, a, "/bogus", transferarchiver.ErrNoSuchDirectory)
	})

	t.Run("archive empty directory", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "tar_archiver_tests_")
		if err != nil {
			t.Fatal(err)
		}

		archive(t, a, dir, transferarchiver.ErrEmptyDirectory)
	})

	t.Run("archive non-empty directory", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "tar_archiver_tests_")
		if err != nil {
			t.Fatal(err)
		}

		if err = os.MkdirAll(filepath.Join(dir, "foo", "bar"), 0777); err != nil {
			t.Fatal(err)
		}

		if err = ioutil.WriteFile(filepath.Join(dir, "foo", "bar", "hello.txt"), []byte("hello, world"), 0700); err != nil {
			t.Fatal(err)
		}

		objs := archive(t, a, dir, nil)
		if len(objs) != 1 {
			t.Fatal("expected exactly one object from tar archiver")
		}

		if len(objs[transferarchiver.TarArchiverKey]) == 0 {
			t.Fatal("created tar bytes should not be empty")
		}

		t.Run("unarchive to non-empty directory", func(t *testing.T) {
			if err := a.Unarchive(dir, rep, func(k string, w io.WriterAt) error {
				_, err := w.WriteAt(objs[transferarchiver.TarArchiverKey], 0)
				return err
			}); err == nil {
				t.Fatal("should error upon encountering a non empty directory for untar")
			}
		})

		t.Run("unarchive to empty directory", func(t *testing.T) {
			tdir, err := ioutil.TempDir("", "tar_unarchive_test")
			if err != nil {
				t.Fatal(err)
			}

			if err = a.Unarchive(tdir, rep, func(k string, w io.WriterAt) error {
				_, err = w.WriteAt(objs[transferarchiver.TarArchiverKey], 0)
				return err
			}); err != nil {
				t.Fatal(err)
			}

			p := filepath.Join(tdir, "foo", "bar", "hello.txt")
			d, err := ioutil.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(d, []byte("hello, world")) {
				t.Fatal("unarchived file content should be equal")
			}

			fi, err := os.Stat(p)
			if err != nil {
				t.Fatal(err)
			}

			if fi.Mode() != 0700 {
				t.Fatalf("expected file permissions to equal what was archived, got: %s", fi.Mode())
			}
		})
	})
}
