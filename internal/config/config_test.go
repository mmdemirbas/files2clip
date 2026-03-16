package config

import (
	"os"
	"path/filepath"
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
}

func TestParseSize(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"1000", 1000, false},
		{"0", 0, false},
		{"10MB", 10_000_000, false},
		{"10mb", 10_000_000, false},
		{"10 MB", 10_000_000, false},
		{"1.5GB", 1_500_000_000, false},
		{"500KB", 500_000, false},
		{"2.5kb", 2500, false},
		{"", 0, true},
		{"-1MB", 0, true},
		{"abc", 0, true},
		{"MB", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
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
		{500, "500 B"},
		{1500, "1.5 KB"},
		{10_000_000, "10.0 MB"},
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

	t.Run("empty file uses defaults", func(t *testing.T) {
		path := writeTempConfig(t, "")

		cfg, err := LoadFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defaults := DefaultConfig()
		if cfg != defaults {
			t.Errorf("empty config should equal defaults: got %+v, want %+v", cfg, defaults)
		}
	})

	t.Run("comments only", func(t *testing.T) {
		content := "# comment\n# another comment\n"
		path := writeTempConfig(t, content)

		cfg, err := LoadFromFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		defaults := DefaultConfig()
		if cfg != defaults {
			t.Errorf("comments-only config should equal defaults")
		}
	})

	t.Run("invalid syntax", func(t *testing.T) {
		content := "no_equals_sign\n"
		path := writeTempConfig(t, content)

		_, err := LoadFromFile(path)
		if err == nil {
			t.Fatal("expected error for invalid syntax")
		}
	})

	t.Run("invalid size value", func(t *testing.T) {
		content := "max_file_size = notanumber\n"
		path := writeTempConfig(t, content)

		_, err := LoadFromFile(path)
		if err == nil {
			t.Fatal("expected error for invalid size")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := LoadFromFile("/nonexistent/config")
		if err == nil {
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
