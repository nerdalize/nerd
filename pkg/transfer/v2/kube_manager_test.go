package transfer_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	transfer "github.com/nerdalize/nerd/pkg/transfer/v2"
	"github.com/nerdalize/nerd/svc"
)

func testManager(tb testing.TB) (mgr *transfer.KubeManager, clean func()) {
	var err error
	stores := map[transfer.StoreType]transfer.StoreFactory{
		transfer.StoreTypeS3: transfer.CreateS3Store,
	}

	archivers := map[transfer.ArchiverType]transfer.ArchiverFactory{
		transfer.ArchiverTypeTar: transfer.CreateTarArchiver,
	}

	di, cleanNs, err := svc.TempDI("")
	if err != nil {
		tb.Fatal(err)
	}

	kube := svc.NewKube(di)

	mgr, err = transfer.NewKubeManager(kube, stores, archivers)
	if err != nil {
		tb.Fatal(err)
	}

	return mgr, func() {
		cleanNs()
	}
}

func TestKubeManager(t *testing.T) {
	var mgr transfer.Manager
	var clean func()

	mgr, clean = testManager(t)
	defer clean()

	opts, _, clean := testS3Store(t)
	defer clean()

	ctx := context.Background()
	st := transfer.StoreTypeS3
	at := transfer.ArchiverTypeTar

	t.Run("create should succeed on non-existing dataset", func(t *testing.T) {
		h1, err := mgr.Create(ctx, "ds-1", st, at, opts)
		if err != nil {
			t.Fatal(err)
		}

		// if h1.Name() == "" {
		// 	t.Fatal("expected name to be set on handle")
		// }

		t.Run("new create should fail on existing dataset", func(t *testing.T) {
			_, err = mgr.Create(ctx, "ds-1", st, at, opts)
			if err == nil {
				t.Fatal("new create should fail on existing dataset")
			}
		})

		t.Run("re-open should fail on already opened dataset", func(t *testing.T) {
			//@TODO implement a distributed lock/mutex/semaphore for a dataset
			//to satisfy this test case
		})

		err = h1.Close()
		if err != nil {
			t.Fatal(err)
		}

		t.Run("re-open after close should open the dataset without error", func(t *testing.T) {
			_, err = mgr.Open(ctx, "ds-1")
			if err != nil {
				t.Fatal(err)
			}

			// if h2.Name() != h1.Name() {
			// 	t.Fatal("expected re-open to return the same name")
			// }
		})

		t.Run("deleting an existing dataset should work", func(t *testing.T) {
			err = mgr.Remove(ctx, "ds-1")
			if err != nil {
				t.Fatal(err)
			}

			//@TODO this is bad, when remove returns the resource is still there for a bit(?)
			time.Sleep(time.Second)

			t.Run("open should fail on non-existing dataset", func(t *testing.T) {
				_, err := mgr.Open(ctx, "ds-1")
				if err == nil {
					t.Fatal("expected dataset open to fail with error")
				}
			})
		})
	})
}

func TestKubeHandle(t *testing.T) {
	mgr, clean := testManager(t)
	defer clean()

	opts, _, clean := testS3Store(t)
	defer clean()

	ctx := context.Background()
	st := transfer.StoreTypeS3
	at := transfer.ArchiverTypeTar

	t.Run("create should succeed for upload", func(t *testing.T) {
		h1, err := mgr.Create(ctx, "ds-1", st, at, opts)
		if err != nil {
			t.Fatal(err)
		}

		dir, err1 := ioutil.TempDir("", "s3_mgr_test_")
		err2 := os.MkdirAll(filepath.Join(dir, "foo", "bar"), 0777)
		err3 := ioutil.WriteFile(filepath.Join(dir, "foo", "bar", "hello.txt"), []byte("hello, world"), 0700)
		if err1 != nil || err2 != nil || err3 != nil {
			t.Fatal(err1, err2, err3)
		}

		err = h1.Push(ctx, dir, transfer.DiscardReporter())
		if err != nil {
			t.Fatal(err)
		}

		dir2, err := ioutil.TempDir("", "s3_mgr_test_")
		if err != nil {
			t.Fatal(err)
		}

		err = h1.Pull(ctx, dir2, transfer.DiscardReporter())
		if err != nil {
			t.Fatal(err)
		}

		size, err := mgr.Info(ctx, "ds-1")
		if err != nil {
			t.Fatal(err)
		}

		if size == 0 {
			t.Fatal("file size should be set to non-zero")
		}

		err = h1.Clear(ctx, transfer.DiscardReporter())
		if err != nil {
			t.Fatal(err)
		}

		size, err = mgr.Info(ctx, "ds-1")
		if err != nil {
			t.Fatal(err)
		}

		if size != 0 {
			t.Fatal("expected file size to be zero now")
		}
	})
}
