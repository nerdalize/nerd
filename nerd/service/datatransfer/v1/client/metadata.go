package v1data

import (
	"bytes"
	"context"
	"encoding/json"
	"path"

	"github.com/nerdalize/nerd/nerd/client"
	v1payload "github.com/nerdalize/nerd/nerd/service/datatransfer/v1/client/payload"
)

const (
	//MetadataObjectKey is the key of the object that contains an a dataset's metadata.
	MetadataObjectKey = "metadata"
)

//MetadataExists checks if the metadata object exists.
func (c *Client) MetadataExists(ctx context.Context, bucket, root string) (bool, error) {
	return c.Exists(ctx, bucket, path.Join(root, MetadataObjectKey))
}

//MetadataUpload uploads a dataset's metadata.
func (c *Client) MetadataUpload(ctx context.Context, bucket, root string, m *v1payload.Metadata) error {
	dat, err := json.Marshal(m)
	if err != nil {
		return client.NewError("failed to encode metadata", err)
	}
	err = c.Upload(ctx, bucket, path.Join(root, MetadataObjectKey), bytes.NewReader(dat))
	if err != nil {
		return client.NewError("failed to upload index file", err)
	}
	return nil
}

//MetadataDownload downloads a dataset's metadata.
func (c *Client) MetadataDownload(ctx context.Context, bucket, root string) (*v1payload.Metadata, error) {
	r, err := c.Download(ctx, bucket, path.Join(root, MetadataObjectKey))
	if err != nil {
		return nil, client.NewError("failed to download metadata", err)
	}
	defer r.Close()
	dec := json.NewDecoder(r)
	metadata := &v1payload.Metadata{}
	err = dec.Decode(metadata)
	if err != nil {
		return nil, client.NewError("failed to decode metadata", err)
	}
	return metadata, nil
}
