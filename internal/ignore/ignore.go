package ignore

import (
	"os"
	"path/filepath"
	"strings"
)

// Matcher checks paths against gitignore-style patterns.
type Matcher struct {
	patterns []pattern
}

type pattern struct {
	negated  bool
	dirOnly  bool
	anchored bool   // pattern contains a slash (match full path, not just basename)
	raw      string // original pattern for debugging
}

// LoadFile reads a gitignore-style file and returns a Matcher.
// Supported syntax:
//   - Lines starting with # are comments
//   - Lines starting with ! negate a previous match
//   - Trailing / matches only directories
//   - Patterns without / match the basename only
//   - Patterns with / match the full relative path
//   - * matches anything except /
//   - ? matches any single character except /
//   - ** matches everything including path separators
func LoadFile(path string) (*Matcher, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(string(data)), nil
}

// Merge combines two Matchers into one. Patterns from b are appended after a.
func Merge(a, b *Matcher) *Matcher {
	merged := &Matcher{}
	if a != nil {
		merged.patterns = append(merged.patterns, a.patterns...)
	}
	if b != nil {
		merged.patterns = append(merged.patterns, b.patterns...)
	}
	return merged
}

// Parse creates a Matcher from gitignore-style pattern text.
func Parse(text string) *Matcher {
	var patterns []pattern
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		p := pattern{raw: line}

		if strings.HasPrefix(line, "!") {
			p.negated = true
			line = line[1:]
		}

		if strings.HasSuffix(line, "/") {
			p.dirOnly = true
			line = strings.TrimRight(line, "/")
		}

		// A leading / anchors to the root but is not part of the match pattern
		if strings.HasPrefix(line, "/") {
			p.anchored = true
			line = line[1:]
		}

		// If the pattern contains a slash, it's anchored to the path structure
		if strings.Contains(line, "/") {
			p.anchored = true
		}

		p.raw = line
		patterns = append(patterns, p)
	}
	return &Matcher{patterns: patterns}
}

// Match reports whether the given path should be ignored.
// The path should be slash-separated and relative (e.g., "src/main.go").
// isDir indicates whether the path is a directory.
func (m *Matcher) Match(path string, isDir bool) bool {
	if m == nil || len(m.patterns) == 0 {
		return false
	}

	// Normalize to forward slashes
	path = strings.TrimPrefix(filepath.ToSlash(path), "/")

	matched := false
	for _, p := range m.patterns {
		if p.dirOnly && !isDir {
			continue
		}

		var hit bool
		if p.anchored {
			hit = globMatch(p.raw, path)
		} else {
			// Match against basename
			base := filepath.Base(path)
			hit = globMatch(p.raw, base)
			// Also try matching against the full path
			if !hit {
				hit = globMatch(p.raw, path)
			}
		}

		if hit {
			matched = !p.negated
		}
	}
	return matched
}

// globMatch matches a gitignore-style pattern against a path.
// Supports *, ?, and ** (which matches across path separators).
func globMatch(pattern, path string) bool {
	return doMatch(pattern, path)
}

func doMatch(pattern, name string) bool {
	for len(pattern) > 0 {
		switch {
		case strings.HasPrefix(pattern, "**"):
			// ** matches zero or more path components
			pattern = strings.TrimPrefix(pattern[2:], "/")
			if pattern == "" {
				return true
			}
			// Try matching the rest of pattern against every suffix of name
			for i := 0; i <= len(name); i++ {
				if doMatch(pattern, name[i:]) {
					return true
				}
			}
			return false

		case pattern[0] == '*':
			// * matches everything except /
			pattern = pattern[1:]
			if pattern == "" {
				return !strings.Contains(name, "/")
			}
			for i := 0; i <= len(name); i++ {
				if i > 0 && name[i-1] == '/' {
					break
				}
				if doMatch(pattern, name[i:]) {
					return true
				}
			}
			return false

		case pattern[0] == '?':
			if len(name) == 0 || name[0] == '/' {
				return false
			}
			pattern = pattern[1:]
			name = name[1:]

		default:
			if len(name) == 0 || pattern[0] != name[0] {
				return false
			}
			pattern = pattern[1:]
			name = name[1:]
		}
	}
	return len(name) == 0
}
