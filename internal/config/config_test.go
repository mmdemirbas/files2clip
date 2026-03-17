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
	t.Run("full config", func(t *testing.T) {
		content := "# files2clip config\nmax_file_size = 5MB\nmax_total_size = 20MB\nmax_files = 500\n"
		path := writeTempConfig(t, content)

		cfg, err := LoadFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.MaxFileSize != 5_000_000 {
			t.Errorf("MaxFileSize = %d, want 5000000", cfg.MaxFileSize)
		}
		if cfg.MaxTotalSize != 20_000_000 {
			t.Errorf("MaxTotalSize = %d, want 20000000", cfg.MaxTotalSize)
		}
		if cfg.MaxFiles != 500 {
			t.Errorf("MaxFiles = %d, want 500", cfg.MaxFiles)
		}
		// Unset fields should keep their defaults
		if cfg.FullPaths {
			t.Error("FullPaths should remain default false")
		}
		if cfg.IgnoreFile != "" {
			t.Errorf("IgnoreFile should remain default empty, got %q", cfg.IgnoreFile)
		}
	})

	t.Run("partial config uses defaults", func(t *testing.T) {
		content := "max_files = 42\n"
		path := writeTempConfig(t, content)

		cfg, err := LoadFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.MaxFiles != 42 {
			t.Errorf("MaxFiles = %d, want 42", cfg.MaxFiles)
		}
		if cfg.MaxFileSize != 10_000_000 {
			t.Errorf("MaxFileSize should be default 10000000, got %d", cfg.MaxFileSize)
		}
	})

	t.Run("full_paths and ignore_file", func(t *testing.T) {
		content := "full_paths = true\nignore_file = ~/.config/files2clip/ignore\n"
		path := writeTempConfig(t, content)

		cfg, err := LoadFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cfg.FullPaths {
			t.Error("FullPaths should be true")
		}
		if cfg.IgnoreFile != "~/.config/files2clip/ignore" {
			t.Errorf("IgnoreFile = %q, want %q", cfg.IgnoreFile, "~/.config/files2clip/ignore")
		}
	})

	t.Run("full_paths variants", func(t *testing.T) {
		for _, tc := range []struct {
			value string
			want  bool
		}{
			{"true", true},
			{"yes", true},
			{"1", true},
			{"TRUE", true},
			{"Yes", true},
			{"false", false},
			{"no", false},
			{"0", false},
			{"anything", false},
		} {
			t.Run("full_paths="+tc.value, func(t *testing.T) {
				content := "full_paths = " + tc.value + "\n"
				path := writeTempConfig(t, content)

				cfg, err := LoadFromFile(path)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if cfg.FullPaths != tc.want {
					t.Errorf("got %v, want %v", cfg.FullPaths, tc.want)
				}
			})
		}
	})

	t.Run("empty file uses defaults", func(t *testing.T) {
		path := writeTempConfig(t, "")

		cfg, err := LoadFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg != DefaultConfig() {
			t.Errorf("empty config should equal defaults: got %+v, want %+v", cfg, DefaultConfig())
		}
	})

	t.Run("comments only", func(t *testing.T) {
		content := "# comment\n# another comment\n"
		path := writeTempConfig(t, content)

		cfg, err := LoadFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg != DefaultConfig() {
			t.Error("comments-only config should equal defaults")
		}
	})

	t.Run("invalid syntax", func(t *testing.T) {
		path := writeTempConfig(t, "no_equals_sign\n")
		if _, err := LoadFromFile(path); err == nil {
			t.Fatal("expected error for invalid syntax")
		}
	})

	t.Run("invalid max_file_size", func(t *testing.T) {
		path := writeTempConfig(t, "max_file_size = notanumber\n")
		if _, err := LoadFromFile(path); err == nil {
			t.Fatal("expected error for invalid max_file_size")
		}
	})

	t.Run("invalid max_total_size", func(t *testing.T) {
		path := writeTempConfig(t, "max_total_size = notanumber\n")
		if _, err := LoadFromFile(path); err == nil {
			t.Fatal("expected error for invalid max_total_size")
		}
	})

	t.Run("invalid max_files", func(t *testing.T) {
		path := writeTempConfig(t, "max_files = notanumber\n")
		if _, err := LoadFromFile(path); err == nil {
			t.Fatal("expected error for invalid max_files")
		}
	})

	t.Run("unknown key does not error", func(t *testing.T) {
		path := writeTempConfig(t, "unknown_key = value\n")

		cfg, err := LoadFromFile(path)
		if err != nil {
			t.Fatalf("unknown key should not cause error: %v", err)
		}
		if cfg != DefaultConfig() {
			t.Errorf("unknown key should not change defaults: got %+v", cfg)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		if _, err := LoadFromFile("/nonexistent/config"); err == nil {
			t.Fatal("expected error for nonexistent file")
		}
	})

	t.Run("windows line endings", func(t *testing.T) {
		content := "max_files = 99\r\nmax_file_size = 1MB\r\n"
		path := writeTempConfig(t, content)

		cfg, err := LoadFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.MaxFiles != 99 {
			t.Errorf("MaxFiles = %d, want 99", cfg.MaxFiles)
		}
		if cfg.MaxFileSize != 1_000_000 {
			t.Errorf("MaxFileSize = %d, want 1000000", cfg.MaxFileSize)
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

func BenchmarkParseSize(b *testing.B) {
	for b.Loop() {
		ParseSize("10MB")
	}
}

func BenchmarkLoadFromFile(b *testing.B) {
	content := "max_file_size = 5MB\nmax_total_size = 20MB\nmax_files = 500\n"
	dir := b.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		LoadFromFile(path)
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
	return path
}
