package pathutil

import (
	"os"
	"path/filepath"
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
	t.Run("normal file with paths", func(t *testing.T) {
		content := "/path/to/file1.txt\n/path/to/file2.txt\n/path/to/dir\n"
		path := writeTempFile(t, content)

		paths, err := ReadPathsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(paths) != 3 {
			t.Fatalf("expected 3 paths, got %d", len(paths))
		}
		if paths[0] != "/path/to/file1.txt" {
			t.Errorf("paths[0] = %q, want %q", paths[0], "/path/to/file1.txt")
		}
	})

	t.Run("file with comments and empty lines", func(t *testing.T) {
		content := "# This is a comment\n\n/path/to/file.txt\n\n# Another comment\n/path/to/other.txt\n"
		path := writeTempFile(t, content)

		paths, err := ReadPathsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(paths) != 2 {
			t.Fatalf("expected 2 paths, got %d", len(paths))
		}
	})

	t.Run("empty file", func(t *testing.T) {
		path := writeTempFile(t, "")

		paths, err := ReadPathsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(paths) != 0 {
			t.Fatalf("expected 0 paths, got %d", len(paths))
		}
	})

	t.Run("file with only comments", func(t *testing.T) {
		content := "# comment 1\n# comment 2\n"
		path := writeTempFile(t, content)

		paths, err := ReadPathsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(paths) != 0 {
			t.Fatalf("expected 0 paths, got %d", len(paths))
		}
	})

	t.Run("whitespace trimming", func(t *testing.T) {
		content := "  /path/to/file.txt  \n\t/path/to/other.txt\t\n"
		path := writeTempFile(t, content)

		paths, err := ReadPathsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(paths) != 2 {
			t.Fatalf("expected 2 paths, got %d", len(paths))
		}
		if paths[0] != "/path/to/file.txt" {
			t.Errorf("paths[0] = %q, want %q", paths[0], "/path/to/file.txt")
		}
		if paths[1] != "/path/to/other.txt" {
			t.Errorf("paths[1] = %q, want %q", paths[1], "/path/to/other.txt")
		}
	})

	t.Run("windows line endings", func(t *testing.T) {
		content := "/path/to/file1.txt\r\n/path/to/file2.txt\r\n"
		path := writeTempFile(t, content)

		paths, err := ReadPathsFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(paths) != 2 {
			t.Fatalf("expected 2 paths, got %d", len(paths))
		}
		if paths[0] != "/path/to/file1.txt" {
			t.Errorf("paths[0] = %q, want %q", paths[0], "/path/to/file1.txt")
		}
		if paths[1] != "/path/to/file2.txt" {
			t.Errorf("paths[1] = %q, want %q", paths[1], "/path/to/file2.txt")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := ReadPathsFromFile("/nonexistent/file.txt")
		if err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})
}

func TestParsePaths(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
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
			if len(got) != len(tt.expected) {
				t.Fatalf("ParsePaths() returned %d paths, want %d", len(got), len(tt.expected))
			}
			for i, p := range got {
				if p != tt.expected[i] {
					t.Errorf("ParsePaths()[%d] = %q, want %q", i, p, tt.expected[i])
				}
			}
		})
	}
}

func BenchmarkReadPathsFromFile(b *testing.B) {
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, "/some/path/to/file"+string(rune('0'+i%10))+".go")
	}
	content := ""
	for _, l := range lines {
		content += l + "\n"
	}
	dir := b.TempDir()
	path := filepath.Join(dir, "paths.txt")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		ReadPathsFromFile(path)
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
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}
