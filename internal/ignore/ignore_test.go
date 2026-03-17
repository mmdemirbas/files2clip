package ignore

import "testing"

func TestMatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		isDir   bool
		want    bool
	}{
		// Simple filename patterns
		{"match extension", "*.log", "debug.log", false, true},
		{"match extension nested", "*.log", "src/debug.log", false, true},
		{"no match extension", "*.log", "debug.txt", false, false},

		// Directory patterns
		{"dir pattern on dir", "build/", "build", true, true},
		{"dir pattern on file", "build/", "build", false, false},
		{"dir pattern nested", "node_modules/", "node_modules", true, true},

		// Anchored patterns (contain /)
		{"anchored match", "src/*.go", "src/main.go", false, true},
		{"anchored no match", "src/*.go", "other/main.go", false, false},

		// ** patterns
		{"doublestar prefix", "**/foo", "foo", false, true},
		{"doublestar prefix nested", "**/foo", "a/b/foo", false, true},
		{"doublestar middle", "a/**/z", "a/z", false, true},
		{"doublestar middle deep", "a/**/z", "a/b/c/z", false, true},
		{"doublestar suffix", "src/**", "src/main.go", false, true},
		{"doublestar suffix deep", "src/**", "src/a/b/c.go", false, true},

		// Negation
		{"negation", "*.log\n!important.log", "debug.log", false, true},
		{"negation keeps", "*.log\n!important.log", "important.log", false, false},

		// Leading slash anchor
		{"leading slash", "/build", "build", false, true},
		{"leading slash no deep match", "/build", "src/build", false, false},

		// Comments and blank lines
		{"comment ignored", "# comment\n*.log", "debug.log", false, true},
		{"blank lines ignored", "\n\n*.log\n\n", "debug.log", false, true},

		// Question mark
		{"question mark", "file?.txt", "file1.txt", false, true},
		{"question mark no slash", "file?.txt", "file/.txt", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Parse(tt.pattern)
			got := m.Match(tt.path, tt.isDir)
			if got != tt.want {
				t.Errorf("Match(%q, isDir=%v) with pattern %q = %v, want %v",
					tt.path, tt.isDir, tt.pattern, got, tt.want)
			}
		})
	}
}

func TestNilMatcher(t *testing.T) {
	var m *Matcher
	if m.Match("foo.txt", false) {
		t.Error("nil matcher should not match anything")
	}
}

func BenchmarkMatch(b *testing.B) {
	m := Parse("*.log\nnode_modules/\nsrc/**/*.test.go\n!important.log")
	for b.Loop() {
		m.Match("src/pkg/handler_test.go", false)
	}
}
