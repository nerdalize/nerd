package v1data

import (
	"bytes"
	"testing"
)

func TestIndexReader(t *testing.T) {
	keys := []Key{
		Key{1, 2, 3},
		Key{4, 5, 6},
		Key{7, 8, 9},
	}
	// create input stream
	buf := bytes.NewBufferString("")
	for _, key := range keys {
		buf.WriteString(key.ToString() + "\n")
	}

	// new index reader
	ir := NewIndexReader(buf)

	// compare input with reader result
	for _, key := range keys {
		read, err := ir.ReadKey()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if key.ToString() != read.ToString() {
			t.Errorf("expected key %v but got %v", key, read)
		}
	}
}

func TestIndexWriter(t *testing.T) {
	keys := []Key{
		Key{1, 2, 3},
		Key{4, 5, 6},
		Key{7, 8, 9},
	}
	// create input stream
	expected := bytes.NewBufferString("")
	for _, key := range keys {
		expected.WriteString(key.ToString() + "\n")
	}
	result := bytes.NewBuffer(nil)
	iw := NewIndexWriter(result)
	for _, key := range keys {
		err := iw.WriteKey(key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if expected.String() != result.String() {
		t.Errorf("expected %v but got %v", expected, result)
	}
}
