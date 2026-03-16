package pathutil

import "testing"

func TestCommonDir(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		expected string
	}{
		{
			name:     "same directory",
			paths:    []string{"/home/user/file1.txt", "/home/user/file2.txt", "/home/user/file3.txt"},
			expected: "/home/user/",
		},
		{
			name:     "different directories with common directory",
			paths:    []string{"/home/user/file1.txt", "/home/user/file2.txt", "/home/user2/file3.txt"},
			expected: "/home/",
		},
		{
			name:     "different directories with no common directory",
			paths:    []string{"/home/user/file1.txt", "/home/user2/file2.txt", "/root/file3.txt"},
			expected: "/",
		},
		{
			name:     "empty paths",
			paths:    []string{},
			expected: "",
		},
		{
			name:     "single path",
			paths:    []string{"/home/user/file1.txt"},
			expected: "/home/user/",
		},
		{
			name:     "root directory",
			paths:    []string{"/file1.txt", "/file2.txt", "/file3.txt"},
			expected: "/",
		},
		{
			name:     "windows paths",
			paths:    []string{"C:\\Users\\user\\file1.txt", "C:\\Users\\user\\file2.txt"},
			expected: "C:\\Users\\user\\",
		},
		{
			name:     "deeply nested common path",
			paths:    []string{"/a/b/c/d/e/f.txt", "/a/b/c/d/e/g.txt"},
			expected: "/a/b/c/d/e/",
		},
		{
			name:     "one path is prefix of another directory name",
			paths:    []string{"/home/user/file.txt", "/home/username/file.txt"},
			expected: "/home/",
		},
		{
			name:     "identical paths",
			paths:    []string{"/home/user/file.txt", "/home/user/file.txt"},
			expected: "/home/user/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := CommonDir(tt.paths)
			if actual != tt.expected {
				t.Errorf("expected: %q, got: %q", tt.expected, actual)
			}
		})
	}
}

func BenchmarkCommonDir(b *testing.B) {
	paths := []string{
		"/home/user/projects/myapp/src/main.go",
		"/home/user/projects/myapp/src/util.go",
		"/home/user/projects/myapp/src/handler.go",
		"/home/user/projects/myapp/README.md",
		"/home/user/projects/myapp/go.mod",
	}
	for b.Loop() {
		CommonDir(paths)
	}
}

func BenchmarkCommonDirSinglePath(b *testing.B) {
	paths := []string{"/home/user/projects/myapp/src/main.go"}
	for b.Loop() {
		CommonDir(paths)
	}
}

func BenchmarkCommonDirManyPaths(b *testing.B) {
	paths := make([]string, 100)
	for i := range paths {
		paths[i] = "/home/user/projects/myapp/src/pkg/internal/deep/file" + string(rune('0'+i%10)) + ".go"
	}
	for b.Loop() {
		CommonDir(paths)
	}
}
