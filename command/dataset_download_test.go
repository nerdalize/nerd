package command

import (
	"os"
	"path"
	"testing"
)

func TestSafeFilePath(t *testing.T) {
	tmpPrefix := "download_test"
	cases := []struct {
		present  []string
		input    string
		expected string
	}{
		{
			present:  []string{},
			input:    "filename.txt",
			expected: "filename.txt",
		},
		{
			present:  []string{"filename.txt"},
			input:    "filename.txt",
			expected: "filename_(1).txt",
		},
		{
			present:  []string{"filename.txt", "filename_(1).txt"},
			input:    "filename.txt",
			expected: "filename_(2).txt",
		},
		{
			present:  []string{"filename.txt", "filename_(1).txt", "filename_(2).txt"},
			input:    "filename.txt",
			expected: "filename_(3).txt",
		},
		{
			present:  []string{"filename.txt", "filename_(1).txt", "filename_(2).txt"},
			input:    "filename_(2).txt",
			expected: "filename_(3).txt",
		},
		{
			present:  []string{"filename.txt", "filename_(2).txt"},
			input:    "filename_(1).txt",
			expected: "filename_(1).txt",
		},
		{
			present:  []string{"filename.dot.txt"},
			input:    "filename.txt",
			expected: "filename.txt",
		},
		{
			present:  []string{"filename"},
			input:    "filename",
			expected: "filename_(1)",
		},
		{
			present:  []string{"filename ()_(1)"},
			input:    "filename ()_(1)",
			expected: "filename ()_(2)",
		},
		{
			present:  []string{"filename(1)"},
			input:    "filename(1)",
			expected: "filename(1)_(1)",
		},
		{
			present:  []string{"filename (1)"},
			input:    "filename (1)",
			expected: "filename (1)_(1)",
		},
		{
			present:  []string{"filename(1).ext"},
			input:    "filename(1).ext",
			expected: "filename(1)_(1).ext",
		},
	}
	for _, tc := range cases {
		tmp := path.Join(os.TempDir(), tmpPrefix)
		os.Mkdir(tmp, 0777)
		// create files
		for _, file := range tc.present {
			fi, err := os.Create(path.Join(tmp, file))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			err = fi.Close()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}
		// test
		safe := safeFilePath(path.Join(tmp, tc.input))
		expected := path.Join(tmp, tc.expected)
		if safe != expected {
			t.Errorf("Expected %v but got %v, for test case {%v, %v}", expected, safe, tc.present, tc.input)
		}
		// clean up files
		for _, file := range tc.present {
			err := os.Remove(path.Join(tmp, file))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}
	}
}
