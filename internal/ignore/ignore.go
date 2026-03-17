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
	anchored bool     // match against full path, not just basename
	parts    []string // pattern split by "/" (with "**" preserved as-is)
}

// LoadFile reads a gitignore-style file and returns a Matcher.
//
// Supported syntax (following git's gitignore specification):
//   - Blank lines are ignored
//   - Lines starting with # are comments
//   - Leading \ escapes # or ! (e.g., \#file matches file named #file)
//   - Leading ! negates the pattern (re-includes a previously excluded path)
//   - Trailing / matches only directories
//   - Leading / or a / in the middle anchors the pattern to the path root
//   - Patterns without / (other than trailing) match the basename only
//   - * matches anything except /
//   - ? matches any single character except /
//   - [...] matches one character in the set (supports ranges like [a-z])
//   - [!...] or [^...] matches one character NOT in the set
//   - ** when adjacent to / matches zero or more directories:
//   - **/foo matches foo in any directory
//   - foo/** matches everything inside foo
//   - a/**/b matches a/b, a/x/b, a/x/y/b, etc.
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
		if line == "" || line == "#" {
			continue
		}

		// Handle comments and escape/negation
		if strings.HasPrefix(line, "#") {
			continue
		}

		p := pattern{}

		if strings.HasPrefix(line, "\\#") || strings.HasPrefix(line, "\\!") {
			// Backslash escapes a leading # or ! — treat literally
			line = line[1:]
		} else if strings.HasPrefix(line, "!") {
			p.negated = true
			line = line[1:]
		}

		if strings.HasSuffix(line, "/") {
			p.dirOnly = true
			line = strings.TrimRight(line, "/")
		}

		// A leading / anchors to the root but is stripped from the pattern
		if strings.HasPrefix(line, "/") {
			p.anchored = true
			line = line[1:]
		}

		// If the pattern contains a / (after stripping leading/trailing),
		// it's anchored to the path structure
		if strings.Contains(line, "/") {
			p.anchored = true
		}

		p.parts = strings.Split(line, "/")
		patterns = append(patterns, p)
	}
	return &Matcher{patterns: patterns}
}

// Match reports whether the given path should be ignored.
// The path should use forward slashes and be relative (e.g., "src/main.go").
// isDir indicates whether the path is a directory.
func (m *Matcher) Match(path string, isDir bool) bool {
	if m == nil || len(m.patterns) == 0 {
		return false
	}

	// Normalize to forward slashes and strip leading /
	path = strings.TrimPrefix(filepath.ToSlash(path), "/")

	matched := false
	for _, p := range m.patterns {
		if p.dirOnly && !isDir {
			continue
		}

		var hit bool
		if p.anchored {
			hit = matchPath(p.parts, strings.Split(path, "/"))
		} else {
			// Unanchored: try matching against basename first
			base := filepath.Base(path)
			if len(p.parts) == 1 {
				hit = matchComponent(p.parts[0], base)
			}
			// Also try matching against the full path
			if !hit {
				hit = matchPath(p.parts, strings.Split(path, "/"))
			}
		}

		if hit {
			matched = !p.negated
		}
	}
	return matched
}

// matchPath matches pattern parts against path parts.
func matchPath(patParts, nameParts []string) bool {
	return doMatchParts(patParts, nameParts)
}

func doMatchParts(patParts, nameParts []string) bool {
	for len(patParts) > 0 {
		pat := patParts[0]

		if pat == "**" {
			patParts = patParts[1:]
			if len(patParts) == 0 {
				return true // trailing ** matches everything
			}
			// Try matching the remaining pattern against every suffix of nameParts
			for i := 0; i <= len(nameParts); i++ {
				if doMatchParts(patParts, nameParts[i:]) {
					return true
				}
			}
			return false
		}

		if len(nameParts) == 0 {
			return false
		}

		if !matchComponent(pat, nameParts[0]) {
			return false
		}

		patParts = patParts[1:]
		nameParts = nameParts[1:]
	}
	return len(nameParts) == 0
}

// matchComponent matches a single path component (no slashes).
// Uses filepath.Match which supports *, ?, [...], and \ escaping.
// Converts gitignore's [!...] negation to Go's [^...] syntax.
func matchComponent(pattern, name string) bool {
	ok, err := filepath.Match(convertCharClassNeg(pattern), name)
	if err != nil {
		return false // malformed pattern
	}
	return ok
}

// convertCharClassNeg converts gitignore [!...] negation to Go's [^...].
func convertCharClassNeg(pattern string) string {
	if !strings.Contains(pattern, "[!") {
		return pattern
	}
	var b strings.Builder
	b.Grow(len(pattern))
	for i := 0; i < len(pattern); i++ {
		if pattern[i] == '[' && i+1 < len(pattern) && pattern[i+1] == '!' {
			b.WriteByte('[')
			b.WriteByte('^')
			i++ // skip the !
		} else {
			b.WriteByte(pattern[i])
		}
	}
	return b.String()
}
