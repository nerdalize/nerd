package transferarchiver_test

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	transfer "github.com/nerdalize/nerd/pkg/transfer"
	"github.com/nerdalize/nerd/pkg/transfer/archiver"
)

func archive(tb testing.TB, a transfer.Archiver, dir string) map[string][]byte {
	tb.Helper()

	objs := map[string][]byte{}
	if err := a.Archive(dir, func(k string, r io.Reader) error {
		buf := bytes.NewBuffer(nil)
		_, err := io.Copy(buf, r)
		objs[k] = buf.Bytes()
		return err
	}); err != nil {
		tb.Fatal(err)
	}

	return objs
}

func TestTarArchiver(t *testing.T) {
	var (
		a   transfer.Archiver
		err error
	)

	a, err = transferarchiver.NewTarArchiver(transferarchiver.ArchiverOptions{})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("archive empty directory", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "tar_archiver_tests_")
		if err != nil {
			t.Fatal(err)
		}

		objs := archive(t, a, dir)
		if len(objs) != 1 {
			t.Fatal("expected exactly one object from tar archiver")
		}

		if len(objs[transferarchiver.TarArchiverKey]) != 0 {
			t.Fatal("created tar bytes should be empty")
		}

		t.Run("unarchive to empty directory", func(t *testing.T) {
			if err := a.Unarchive(dir, func(k string, w io.WriterAt) error {
				_, err := w.WriteAt(objs[transferarchiver.TarArchiverKey], 0)
				return err
			}); err != nil {
				t.Fatal(err)
			}

			if err := filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
				if p == dir {
					return nil
				}

				return errors.New("unarching an empty archive should result into a empty dir")
			}); err != nil {
				t.Fatal(err)
			}
		})
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

		objs := archive(t, a, dir)
		if len(objs) != 1 {
			t.Fatal("expected exactly one object from tar archiver")
		}

		if len(objs[transferarchiver.TarArchiverKey]) == 0 {
			t.Fatal("created tar bytes should not be empty")
		}

		t.Run("unarchive to non-empty directory", func(t *testing.T) {
			if err := a.Unarchive(dir, func(k string, w io.WriterAt) error {
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

			if err = a.Unarchive(tdir, func(k string, w io.WriterAt) error {
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
