package transfer_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	transfer "github.com/nerdalize/nerd/pkg/transfer/v2"
)

func testS3Store(tb testing.TB) (opts map[string]string, store transfer.Store, clean func()) {
	if os.Getenv("AWS_ACCESS_KEY_ID") == "" || os.Getenv("AWS_REGION") == "" {
		tb.Skip("must have configured AWS_ACCESS_KEY or AWS_REGION env variable")
	}

	name, cleanBucket, err := transfer.TempS3Bucket()
	if err != nil {
		tb.Fatal(err)
	}

	opts = map[string]string{
		"aws_s3_region":     os.Getenv("AWS_REGION"),
		"aws_s3_bucket":     name,
		"aws_s3_access_key": os.Getenv("AWS_ACCESS_KEY_ID"),
		"aws_s3_secret_key": os.Getenv("AWS_SECRET_ACCESS_KEY"),
		"aws_s3_prefix":     "tests/",
	}

	store, err = transfer.CreateS3Store(opts)
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
		buf1 := bytes.NewBufferString(content1)

		err := store.Put(ctx, "hello.txt", buf1)
		if err != nil {
			t.Fatal(err)
		}

		content2 := "hello, world2"
		t.Run("putting an existing key", func(t *testing.T) {
			buf2 := bytes.NewBufferString(content2)

			err := store.Put(ctx, "hello.txt", buf2)
			if err != nil {
				t.Fatal(err)
			}

			t.Run("get an existing key", func(t *testing.T) {
				buf2 := aws.NewWriteAtBuffer(nil)

				err := store.Get(ctx, "hello.txt", buf2)
				if err != nil {
					t.Fatal(err)
				}

				if !bytes.Equal([]byte(content2), buf2.Bytes()) {
					t.Fatalf("expected downloaded content to equal reuploaded content but got: '%x' vs '%x'", buf1.Bytes(), buf2.Bytes())
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
