package v1data

import (
	"crypto/sha256"
	"strings"
	"testing"
	"time"
)

func TestStreaming(t *testing.T) {
	expected := `{"created_at":"2017-03-27T12:19:27.899270493+02:00","updated_at":"2017-03-27T12:19:27.899270568+02:00","size":43008}
0f9ac64d972a6be4340519ec347de01449a5ff19db42b926a809eb0bafc60427
084d3b6dda5e3d9ad9dec04d749245fa3ff1020462cfdc20faea67a6bc866959`

	metadata, err := NewMetadataFromReader(strings.NewReader(expected))
	if err != nil {
		t.Fatalf("Failed to metadata from reader: %v", err)
	}
	str, err := metadata.ToString()
	if err != nil {
		t.Fatalf("Failed convert metadata to string: %v", err)
	}
	if expected != str {
		t.Errorf("expected string %v, but got %v", expected, str)
	}
}

func TestBuffered(t *testing.T) {
	header := &MetadataHeader{
		Size:    10,
		Created: time.Unix(0, 0),
		Updated: time.Unix(10, 0),
	}
	metadata := NewMetadata(header, NewBufferedKeyReadWiter())
	metadata.WriteKey(Key(sha256.Sum256([]byte("Test1")))) // 8a863b145dc6e4ed7ac41c08f7536c476ebac7509e028ed2b49f8bd5a3562b9f (http://www.xorbin.com/tools/sha256-hash-calculator)
	metadata.WriteKey(Key(sha256.Sum256([]byte("Test2")))) // 32e6e1e134f9cc8f14b05925667c118d19244aebce442d6fecd2ac38cdc97649

	expected := `{"created_at":"1970-01-01T01:00:00+01:00","updated_at":"1970-01-01T01:00:10+01:00","size":10}
8a863b145dc6e4ed7ac41c08f7536c476ebac7509e028ed2b49f8bd5a3562b9f
32e6e1e134f9cc8f14b05925667c118d19244aebce442d6fecd2ac38cdc97649`
	str, err := metadata.ToString()
	if err != nil {
		t.Fatalf("Failed convert metadata to string: %v", err)
	}
	if expected != str {
		t.Errorf("expected string %v, but got %v", expected, str)
	}
}
