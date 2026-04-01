package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Config holds runtime limits for files2clip.
type Config struct {
	MaxFileSize   int64  // max bytes per file, 0 = unlimited
	MaxTotalSize  int64  // max bytes total content, 0 = unlimited
	MaxFiles      int    // max number of files, 0 = unlimited
	FullPaths     bool   // use absolute paths instead of relative
	IgnoreFile    string // path to gitignore-style ignore file
	IncludeBinary bool   // include binary files (skipped by default)
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		MaxFileSize:  10 * 1000 * 1000, // 10 MB
		MaxTotalSize: 50 * 1000 * 1000, // 50 MB
		MaxFiles:     1000,
	}
}

// ConfigFilePath returns the platform-appropriate config file path.
func ConfigFilePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "files2clip", "config"), nil
}

// LoadFromFile reads a key=value config file and merges values over the defaults.
// Unknown keys are reported to stderr but do not cause errors.
func LoadFromFile(path string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path) //nolint:gosec // G304: path is the user's own config file, not external input
	if err != nil {
		return cfg, err
	}

	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		key, value, ok := strings.Cut(trimmed, "=")
		if !ok {
			return cfg, fmt.Errorf("line %d: invalid syntax (expected key = value): %s", i+1, trimmed)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "max_file_size":
			n, err := ParseSize(value)
			if err != nil {
				return cfg, fmt.Errorf("line %d: max_file_size: %w", i+1, err)
			}
			cfg.MaxFileSize = n
		case "max_total_size":
			n, err := ParseSize(value)
			if err != nil {
				return cfg, fmt.Errorf("line %d: max_total_size: %w", i+1, err)
			}
			cfg.MaxTotalSize = n
		case "max_files":
			n, err := strconv.Atoi(value)
			if err != nil {
				return cfg, fmt.Errorf("line %d: max_files: %w", i+1, err)
			}
			cfg.MaxFiles = n
		case "full_paths":
			cfg.FullPaths = parseBool(value)
		case "ignore_file":
			cfg.IgnoreFile = value
		case "include_binary":
			cfg.IncludeBinary = parseBool(value)
		default:
			fmt.Fprintf(os.Stderr, "config: unknown key %q on line %d\n", key, i+1)
		}
	}

	return cfg, nil
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "yes" || s == "1"
}

// ParseSize converts a human-readable size string to bytes.
// Supported suffixes (case-insensitive): KB, MB, GB. Plain integers are bytes.
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size value")
	}

	upper := strings.ToUpper(s)
	multiplier := int64(1)
	numStr := s

	switch {
	case strings.HasSuffix(upper, "GB"):
		multiplier = 1000 * 1000 * 1000
		numStr = s[:len(s)-2]
	case strings.HasSuffix(upper, "MB"):
		multiplier = 1000 * 1000
		numStr = s[:len(s)-2]
	case strings.HasSuffix(upper, "KB"):
		multiplier = 1000
		numStr = s[:len(s)-2]
	}

	numStr = strings.TrimSpace(numStr)
	n, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid size %q: %w", s, err)
	}
	if n < 0 {
		return 0, fmt.Errorf("negative size %q", s)
	}

	return int64(n * float64(multiplier)), nil
}

// FormatSize converts bytes to a human-readable string.
func FormatSize(bytes int64) string {
	switch {
	case bytes >= 1000*1000*1000:
		return fmt.Sprintf("%.1f GB", float64(bytes)/(1000*1000*1000))
	case bytes >= 1000*1000:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1000*1000))
	case bytes >= 1000:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1000)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
