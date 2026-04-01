package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MaxFileSize != 10_000_000 {
		t.Errorf("MaxFileSize = %d, want 10000000", cfg.MaxFileSize)
	}
	if cfg.MaxTotalSize != 50_000_000 {
		t.Errorf("MaxTotalSize = %d, want 50000000", cfg.MaxTotalSize)
	}
	if cfg.MaxFiles != 1000 {
		t.Errorf("MaxFiles = %d, want 1000", cfg.MaxFiles)
	}
	if cfg.FullPaths {
		t.Error("FullPaths should default to false")
	}
	if cfg.IgnoreFile != "" {
		t.Errorf("IgnoreFile should default to empty, got %q", cfg.IgnoreFile)
	}
	if cfg.IncludeBinary {
		t.Error("IncludeBinary should default to false")
	}
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		wantErr  bool
	}{
		{"plain bytes", "1000", 1000, false},
		{"zero", "0", 0, false},
		{"megabytes upper", "10MB", 10_000_000, false},
		{"megabytes lower", "10mb", 10_000_000, false},
		{"megabytes with space", "10 MB", 10_000_000, false},
		{"gigabytes fractional", "1.5GB", 1_500_000_000, false},
		{"kilobytes", "500KB", 500_000, false},
		{"kilobytes fractional lower", "2.5kb", 2500, false},
		{"empty string", "", 0, true},
		{"negative", "-1MB", 0, true},
		{"non-numeric", "abc", 0, true},
		{"suffix only", "MB", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseSize(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.expected {
				t.Errorf("ParseSize(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{999, "999 B"},
		{1000, "1.0 KB"},
		{1500, "1.5 KB"},
		{999_999, "1000.0 KB"},
		{1_000_000, "1.0 MB"},
		{10_000_000, "10.0 MB"},
		{999_999_999, "1000.0 MB"},
		{1_000_000_000, "1.0 GB"},
		{1_500_000_000, "1.5 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := FormatSize(tt.bytes)
			if got != tt.expected {
				t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, got, tt.expected)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	d := DefaultConfig()
	trueFullPaths := Config{MaxFileSize: d.MaxFileSize, MaxTotalSize: d.MaxTotalSize, MaxFiles: d.MaxFiles, FullPaths: true}

	for _, tc := range []struct {
		name    string
		content string
		want    Config
		wantErr bool
	}{
		{
			name:    "full config",
			content: "# files2clip config\nmax_file_size = 5MB\nmax_total_size = 20MB\nmax_files = 500\n",
			want:    Config{MaxFileSize: 5_000_000, MaxTotalSize: 20_000_000, MaxFiles: 500},
		},
		{
			name:    "partial config uses defaults",
			content: "max_files = 42\n",
			want:    Config{MaxFileSize: d.MaxFileSize, MaxTotalSize: d.MaxTotalSize, MaxFiles: 42},
		},
		{
			name:    "full_paths and ignore_file",
			content: "full_paths = true\nignore_file = ~/.config/files2clip/ignore\n",
			want:    Config{MaxFileSize: d.MaxFileSize, MaxTotalSize: d.MaxTotalSize, MaxFiles: d.MaxFiles, FullPaths: true, IgnoreFile: "~/.config/files2clip/ignore"},
		},
		{
			name:    "include_binary",
			content: "include_binary = true\n",
			want:    Config{MaxFileSize: d.MaxFileSize, MaxTotalSize: d.MaxTotalSize, MaxFiles: d.MaxFiles, IncludeBinary: true},
		},
		{name: "empty file uses defaults", content: "", want: d},
		{name: "comments only", content: "# comment\n# another comment\n", want: d},
		{name: "unknown key does not error", content: "unknown_key = value\n", want: d},
		{
			name:    "windows line endings",
			content: "max_files = 99\r\nmax_file_size = 1MB\r\n",
			want:    Config{MaxFileSize: 1_000_000, MaxTotalSize: d.MaxTotalSize, MaxFiles: 99},
		},
		// full_paths bool parsing variants
		{name: "full_paths=true", content: "full_paths = true\n", want: trueFullPaths},
		{name: "full_paths=yes", content: "full_paths = yes\n", want: trueFullPaths},
		{name: "full_paths=1", content: "full_paths = 1\n", want: trueFullPaths},
		{name: "full_paths=TRUE", content: "full_paths = TRUE\n", want: trueFullPaths},
		{name: "full_paths=Yes", content: "full_paths = Yes\n", want: trueFullPaths},
		{name: "full_paths=false", content: "full_paths = false\n", want: d},
		{name: "full_paths=no", content: "full_paths = no\n", want: d},
		{name: "full_paths=0", content: "full_paths = 0\n", want: d},
		{name: "full_paths=anything", content: "full_paths = anything\n", want: d},
		// error cases
		{name: "invalid syntax", content: "no_equals_sign\n", wantErr: true},
		{name: "invalid max_file_size", content: "max_file_size = notanumber\n", wantErr: true},
		{name: "invalid max_total_size", content: "max_total_size = notanumber\n", wantErr: true},
		{name: "invalid max_files", content: "max_files = notanumber\n", wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTempConfig(t, tc.content)
			got, err := LoadFromFile(path)
			if (err != nil) != tc.wantErr {
				t.Fatalf("error = %v, wantErr = %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}
			if got != tc.want {
				t.Errorf("\ngot  %+v\nwant %+v", got, tc.want)
			}
		})
	}

	t.Run("nonexistent file", func(t *testing.T) {
		if _, err := LoadFromFile("/nonexistent/config"); err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})
}

func TestConfigFilePath(t *testing.T) {
	path, err := ConfigFilePath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path == "" {
		t.Fatal("expected non-empty path")
	}
	if !strings.HasSuffix(path, filepath.Join("files2clip", "config")) {
		t.Errorf("path %q should end with files2clip/config", path)
	}
}

func BenchmarkFormatSize(b *testing.B) {
	for b.Loop() {
		FormatSize(10_000_000)
	}
}

func BenchmarkParseSize(b *testing.B) {
	for b.Loop() {
		_, _ = ParseSize("10MB")
	}
}

func BenchmarkLoadFromFile(b *testing.B) {
	content := "max_file_size = 5MB\nmax_total_size = 20MB\nmax_files = 500\n"
	dir := b.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		_, _ = LoadFromFile(path)
	}
}

func FuzzParseSize(f *testing.F) {
	f.Add("10MB")
	f.Add("0")
	f.Add("1.5GB")
	f.Add("")
	f.Add("-1MB")
	f.Add("abc")
	f.Add("MB")
	f.Add("  500 KB  ")

	f.Fuzz(func(t *testing.T, input string) {
		_, _ = ParseSize(input) // must not panic
	})
}

func FuzzLoadFromFile(f *testing.F) {
	f.Add("max_file_size = 10MB\nmax_files = 100\n")
	f.Add("# comment\n\nmax_files = 42\n")
	f.Add("")
	f.Add("bad line\n")
	f.Add("unknown = value\n")
	f.Add("full_paths = true\ninclude_binary = yes\n")

	f.Fuzz(func(t *testing.T, content string) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config")
		if err := os.WriteFile(path, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
		_, _ = LoadFromFile(path) // must not panic
	})
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return path
}
