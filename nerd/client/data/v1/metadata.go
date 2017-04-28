package v1data

import (
	"bytes"
	"encoding/json"
	"path"

	"github.com/nerdalize/nerd/nerd/client"
	v1payload "github.com/nerdalize/nerd/nerd/client/data/v1/payload"
)

const (
	//MetadataObjectKey is the key of the object that contains an a dataset's metadata.
	MetadataObjectKey = "metadata"
)

//MetadataExists checks if the metadata object exists.
func (c *Client) MetadataExists(bucket, root string) (bool, error) {
	return c.Exists(bucket, path.Join(root, MetadataObjectKey))
}

//MetadataUpload uploads a dataset's metadata.
func (c *Client) MetadataUpload(bucket, root string, m *v1payload.Metadata) error {
	dat, err := json.Marshal(m)
	if err != nil {
		return &client.Error{"failed to encode metadata", err}
	}
	err = c.Upload(bucket, path.Join(root, MetadataObjectKey), bytes.NewReader(dat))
	if err != nil {
		return &client.Error{"failed to upload index file", err}
	}
	return nil
}

//MetadataDownload downloads a dataset's metadata.
func (c *Client) MetadataDownload(bucket, root string) (*v1payload.Metadata, error) {
	r, err := c.Download(bucket, path.Join(root, MetadataObjectKey))
	if err != nil {
		return nil, &client.Error{"failed to download metadata", err}
	}
	dec := json.NewDecoder(r)
	metadata := &v1payload.Metadata{}
	err = dec.Decode(metadata)
	if err != nil {
		return nil, &client.Error{"failed to decode metadata", err}
	}
	return metadata, nil
}
