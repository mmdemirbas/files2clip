package pathutil

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// ExcludedFileNames contains OS-specific metadata files that should be skipped.
var ExcludedFileNames = []string{
	".DS_Store",   // macOS Finder metadata file
	"Thumbs.db",   // Windows Explorer metadata file
	"desktop.ini", // Windows desktop configuration file
}

// IsExcluded reports whether the given path should be skipped based on its filename.
func IsExcluded(path string) bool {
	return slices.Contains(ExcludedFileNames, filepath.Base(path))
}

// ReadPathsFromFile reads a file containing one path per line.
// Empty lines and lines starting with # are ignored.
func ReadPathsFromFile(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return ParsePaths(string(data)), nil
}

// ParsePaths extracts paths from text content.
// Empty lines and lines starting with # are ignored.
func ParsePaths(text string) []string {
	lines := strings.Split(text, "\n")
	var effective []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			effective = append(effective, trimmed)
		}
	}
	return effective
}
