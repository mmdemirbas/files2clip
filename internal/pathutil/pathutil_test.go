package pathutil

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestIsExcluded(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/some/path/.DS_Store", true},
		{"/some/path/Thumbs.db", true},
		{"/some/path/desktop.ini", true},
		{"/some/path/file.txt", false},
		{"/some/path/main.go", false},
		{".DS_Store", true},
		{"regular.txt", false},
		{"/deep/nested/.DS_Store", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			actual := IsExcluded(tt.path)
			if actual != tt.expected {
				t.Errorf("IsExcluded(%q) = %v, want %v", tt.path, actual, tt.expected)
			}
		})
	}
}

func TestReadPathsFromFile(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "normal file with paths",
			content: "/path/to/file1.txt\n/path/to/file2.txt\n/path/to/dir\n",
			want:    []string{"/path/to/file1.txt", "/path/to/file2.txt", "/path/to/dir"},
		},
		{
			name:    "file with comments and empty lines",
			content: "# This is a comment\n\n/path/to/file.txt\n\n# Another comment\n/path/to/other.txt\n",
			want:    []string{"/path/to/file.txt", "/path/to/other.txt"},
		},
		{
			name:    "empty file",
			content: "",
			want:    nil,
		},
		{
			name:    "file with only comments",
			content: "# comment 1\n# comment 2\n",
			want:    nil,
		},
		{
			name:    "whitespace trimming",
			content: "  /path/to/file.txt  \n\t/path/to/other.txt\t\n",
			want:    []string{"/path/to/file.txt", "/path/to/other.txt"},
		},
		{
			name:    "windows line endings",
			content: "/path/to/file1.txt\r\n/path/to/file2.txt\r\n",
			want:    []string{"/path/to/file1.txt", "/path/to/file2.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempFile(t, tt.content)

			got, err := ReadPathsFromFile(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}

	t.Run("nonexistent file", func(t *testing.T) {
		if _, err := ReadPathsFromFile("/nonexistent/file.txt"); err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})
}

func TestParsePaths(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"simple paths", "/a\n/b\n/c\n", []string{"/a", "/b", "/c"}},
		{"with comments", "# comment\n/a\n# another\n/b\n", []string{"/a", "/b"}},
		{"empty lines", "\n\n/a\n\n/b\n\n", []string{"/a", "/b"}},
		{"whitespace trimming", "  /a  \n\t/b\t\n", []string{"/a", "/b"}},
		{"windows line endings", "/a\r\n/b\r\n", []string{"/a", "/b"}},
		{"empty input", "", nil},
		{"only whitespace", "  \n  \n", nil},
		{"only comments", "# a\n# b\n", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePaths(tt.input)
			if !slices.Equal(got, tt.want) {
				t.Errorf("ParsePaths() = %v, want %v", got, tt.want)
			}
		})
	}
}

func FuzzParsePaths(f *testing.F) {
	f.Add("/a\n/b\n/c\n")
	f.Add("# comment\n/a\n")
	f.Add("")
	f.Add("  \n  \n")
	f.Add("/a\r\n/b\r\n")

	f.Fuzz(func(t *testing.T, input string) {
		// Must not panic
		ParsePaths(input)
	})
}

func BenchmarkParsePaths(b *testing.B) {
	var buf strings.Builder
	for i := range 100 {
		fmt.Fprintf(&buf, "/some/path/to/file%d.go\n", i)
	}
	input := buf.String()

	for b.Loop() {
		ParsePaths(input)
	}
}

func BenchmarkReadPathsFromFile(b *testing.B) {
	var buf strings.Builder
	for i := range 100 {
		fmt.Fprintf(&buf, "/some/path/to/file%d.go\n", i)
	}
	dir := b.TempDir()
	path := filepath.Join(dir, "paths.txt")
	if err := os.WriteFile(path, []byte(buf.String()), 0600); err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		_, _ = ReadPathsFromFile(path)
	}
}

func BenchmarkIsExcluded(b *testing.B) {
	path := "/some/deep/nested/path/to/file.txt"
	for b.Loop() {
		IsExcluded(path)
	}
}

func BenchmarkIsExcludedMatch(b *testing.B) {
	path := "/some/deep/nested/path/.DS_Store"
	for b.Loop() {
		IsExcluded(path)
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}
