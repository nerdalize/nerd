package transferstore_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	transfer "github.com/nerdalize/nerd/pkg/transfer"
	"github.com/nerdalize/nerd/pkg/transfer/store"
)

func testS3Store(tb testing.TB) (opts transferstore.StoreOptions, store transfer.Store, clean func()) {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_REGION") == "" {
		tb.Skip("must have configured AWS_ACCESS_KEY or AWS_REGION env variable")
	}

	name, cleanBucket, err := transferstore.TempS3Bucket()
	if err != nil {
		tb.Fatal(err)
	}

	opts = transferstore.StoreOptions{
		S3StoreBucket:    name,
		S3StoreAWSRegion: os.Getenv("AWS_REGION"),
		S3StoreAccessKey: os.Getenv("AWS_ACCESS_KEY_ID"),
		S3StoreSecretKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		S3StorePrefix:    "tests/",
	}

	store, err = transferstore.NewS3Store(opts)
	if err != nil {
		tb.Fatal(err)
	}

	return opts, store, func() {
		cleanBucket()
	}
}

func TestS3Store(t *testing.T) {
	ctx := context.Background()
	_, store, clean := testS3Store(t)
	defer clean()

	t.Run("put a non-existing key", func(t *testing.T) {
		content1 := "hello, world"
		buf1, _ := ioutil.TempFile("", "archiver_test_1_")
		buf1.Write([]byte(content1))
		defer buf1.Close()

		err := store.Put(ctx, "hello.txt", buf1)
		if err != nil {
			t.Fatal(err)
		}

		content2 := "hello, world2"
		t.Run("putting an existing key", func(t *testing.T) {
			buf2, _ := ioutil.TempFile("", "archiver_test_2_")
			defer buf2.Close()
			buf2.Write([]byte(content2))
			buf2.Seek(0, 0)

			err = store.Put(ctx, "hello.txt", buf2)
			if err != nil {
				t.Fatal(err)
			}

			t.Run("get an existing key", func(t *testing.T) {
				buf3 := aws.NewWriteAtBuffer(nil)

				err := store.Get(ctx, "hello.txt", buf3)
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal([]byte(content2), buf3.Bytes()) {
					t.Fatalf("expected downloaded content to equal reuploaded content but got: %x", buf3.Bytes())
				}
			})

			t.Run("delete an existing key", func(t *testing.T) {
				err := store.Del(ctx, "hello.txt")
				if err != nil {
					t.Fatal(err)
				}

				t.Run("get an non-existing key", func(t *testing.T) {
					buf3 := aws.NewWriteAtBuffer(nil)

					err := store.Get(ctx, "hello.txt", buf3)
					if err == nil {
						t.Fatal("expected error while getting an non-existing key")
					}
				})
			})
		})
	})

}
