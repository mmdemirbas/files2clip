package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		isDir   bool
		want    bool
	}{
		// ── Simple filename patterns ──
		{"match extension", "*.log", "debug.log", false, true},
		{"match extension nested", "*.log", "src/debug.log", false, true},
		{"no match extension", "*.log", "debug.txt", false, false},
		{"exact name", "Makefile", "Makefile", false, true},
		{"exact name nested", "Makefile", "src/Makefile", false, true},

		// ── Directory-only patterns (trailing /) ──
		{"dir pattern on dir", "build/", "build", true, true},
		{"dir pattern on file", "build/", "build", false, false},
		{"dir pattern nested on dir", "node_modules/", "node_modules", true, true},
		{"dir pattern nested match", "vendor/", "src/vendor", true, true},

		// ── Anchored patterns (contain /) ──
		{"anchored match", "src/*.go", "src/main.go", false, true},
		{"anchored no match deeper", "src/*.go", "src/sub/main.go", false, false},
		{"anchored no match other dir", "src/*.go", "other/main.go", false, false},
		{"anchored dir path", "vendor/pkg", "vendor/pkg", false, true},
		{"anchored dir path no partial", "vendor/pkg", "vendor/pkgs", false, false},

		// ── Leading / anchor ──
		{"leading slash", "/build", "build", false, true},
		{"leading slash no deep match", "/build", "src/build", false, false},
		{"leading slash with wildcard", "/*.go", "main.go", false, true},
		{"leading slash wildcard no deep", "/*.go", "src/main.go", false, false},
		{"leading slash dir on dir", "/build/", "build", true, true},
		{"leading slash dir on file", "/build/", "build", false, false},
		{"leading slash dir nested", "/build/", "src/build", true, false},

		// ── ** (double star) patterns ──
		// bare **
		{"bare doublestar", "**", "anything/at/all", false, true},
		{"bare doublestar file", "**", "file.txt", false, true},
		// leading **/
		{"doublestar prefix", "**/foo", "foo", false, true},
		{"doublestar prefix nested", "**/foo", "a/b/foo", false, true},
		{"doublestar prefix with ext", "**/*.go", "main.go", false, true},
		{"doublestar prefix with ext nested", "**/*.go", "a/b/c/main.go", false, true},
		// trailing /**
		{"doublestar suffix", "src/**", "src/main.go", false, true},
		{"doublestar suffix deep", "src/**", "src/a/b/c.go", false, true},
		{"doublestar suffix no match root", "src/**", "main.go", false, false},
		// middle **/
		{"doublestar middle zero dirs", "a/**/z", "a/z", false, true},
		{"doublestar middle one dir", "a/**/z", "a/b/z", false, true},
		{"doublestar middle deep", "a/**/z", "a/b/c/d/z", false, true},
		{"doublestar middle with wildcard", "src/**/*.test.go", "src/pkg/handler.test.go", false, true},
		{"doublestar middle with wildcard deep", "src/**/*.test.go", "src/a/b/c.test.go", false, true},
		{"doublestar no match", "a/**/impossible.xyz", "a/b/c/d.go", false, false},
		// multiple **
		{"multi doublestar", "a/**/b/**/c", "a/x/b/y/c", false, true},
		{"multi doublestar zero", "a/**/b/**/c", "a/b/c", false, true},

		// ** not adjacent to / (treated as two regular *)
		{"star star no sep", "a**.go", "abc.go", false, true},
		{"star star no sep no cross dir", "a**.go", "a/b.go", false, false},

		// ── Negation (!) ──
		{"negation excludes", "*.log\n!important.log", "debug.log", false, true},
		{"negation re-includes", "*.log\n!important.log", "important.log", false, false},
		{"double negation", "*.log\n!important.log\n*.log", "important.log", false, true},
		{"negation with dir", "build/\n!build/release/", "build/release", true, false},
		{"negate dir-only", "build/\n!build/", "build", true, false},

		// ── Character classes [...] ──
		{"char class simple", "*.[oa]", "test.o", false, true},
		{"char class simple 2", "*.[oa]", "test.a", false, true},
		{"char class no match", "*.[oa]", "test.c", false, false},
		{"char class range", "[Tt]est", "Test", false, true},
		{"char class range lower", "[Tt]est", "test", false, true},
		{"char class range no match", "[Tt]est", "Best", false, false},
		{"char class negated bang", "*.[!o]", "test.a", false, true},
		{"char class negated bang no match", "*.[!o]", "test.o", false, false},
		{"char class range a-z", "[a-z].txt", "m.txt", false, true},
		{"char class range a-z no match", "[a-z].txt", "M.txt", false, false},

		// ── Backslash escaping ──
		{"escaped hash", "\\#file", "#file", false, true},
		{"escaped bang", "\\!file", "!file", false, true},

		// ── Comments and blank lines ──
		{"comment ignored", "# comment\n*.log", "debug.log", false, true},
		{"comment only", "# comment", "anything", false, false},
		{"blank lines ignored", "\n\n*.log\n\n", "debug.log", false, true},
		{"windows line endings", "*.log\r\n*.txt", "foo.txt", false, true},

		// ── Question mark ──
		{"question mark", "file?.txt", "file1.txt", false, true},
		{"question mark no match slash", "file?.txt", "file/.txt", false, false},
		{"question mark no match empty", "file?.txt", "file.txt", false, false},

		// ── Edge cases ──
		{"empty path", "*.log", "", false, false},
		{"empty pattern text", "", "anything", false, false},
		{"root path normalized", "*.go", "/main.go", false, true},
		{"absolute path normalized", "*.go", "/src/main.go", false, true},
		{"no match at all", "*.xyz", "file.abc", false, false},
		{"pattern longer than path", "a/b/c/d", "a/b", false, false},
		{"malformed bracket", "[unclosed", "u", false, false},
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

func TestMerge(t *testing.T) {
	t.Run("combine two matchers", func(t *testing.T) {
		a := Parse("*.log")
		b := Parse("*.txt")
		merged := Merge(a, b)

		if !merged.Match("foo.log", false) {
			t.Error("merged should match *.log")
		}
		if !merged.Match("foo.txt", false) {
			t.Error("merged should match *.txt")
		}
		if merged.Match("foo.go", false) {
			t.Error("merged should not match *.go")
		}
	})

	t.Run("nil first", func(t *testing.T) {
		b := Parse("*.log")
		merged := Merge(nil, b)
		if !merged.Match("foo.log", false) {
			t.Error("Merge(nil, b) should match b's patterns")
		}
	})

	t.Run("nil second", func(t *testing.T) {
		a := Parse("*.log")
		merged := Merge(a, nil)
		if !merged.Match("foo.log", false) {
			t.Error("Merge(a, nil) should match a's patterns")
		}
	})

	t.Run("both nil", func(t *testing.T) {
		merged := Merge(nil, nil)
		if merged.Match("foo.log", false) {
			t.Error("Merge(nil, nil) should not match anything")
		}
	})

	t.Run("negation across merge", func(t *testing.T) {
		a := Parse("*.log")
		b := Parse("!important.log")
		merged := Merge(a, b)

		if !merged.Match("debug.log", false) {
			t.Error("debug.log should still be excluded")
		}
		if merged.Match("important.log", false) {
			t.Error("important.log should be re-included by negation from b")
		}
	})
}

func TestNilMatcher(t *testing.T) {
	var m *Matcher
	if m.Match("foo.txt", false) {
		t.Error("nil matcher should not match anything")
	}
}

func TestLoadFile(t *testing.T) {
	t.Run("nonexistent", func(t *testing.T) {
		_, err := LoadFile("/nonexistent/ignore")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})

	t.Run("valid file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "ignore")
		if err := os.WriteFile(path, []byte("*.log\nnode_modules/\n"), 0600); err != nil {
			t.Fatal(err)
		}

		m, err := LoadFile(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !m.Match("debug.log", false) {
			t.Error("should match *.log")
		}
		if !m.Match("node_modules", true) {
			t.Error("should match node_modules/")
		}
		if m.Match("main.go", false) {
			t.Error("should not match main.go")
		}
	})
}

func FuzzMatch(f *testing.F) {
	f.Add("*.log", "debug.log", false)
	f.Add("src/**/*.go", "src/main.go", false)
	f.Add("build/", "build", true)
	f.Add("!important.log", "important.log", false)
	f.Add("*.[!o]", "test.a", false)
	f.Add("[unclosed", "u", false)
	f.Add("**", "a/b/c", false)
	f.Add("", "", false)

	f.Fuzz(func(t *testing.T, pattern, path string, isDir bool) {
		m := Parse(pattern)
		// Must not panic
		m.Match(path, isDir)
	})
}

func BenchmarkParse(b *testing.B) {
	text := "*.log\nnode_modules/\nsrc/**/*.test.go\n!important.log\n*.o\n*.a\nbuild/\ndist/\n"
	for b.Loop() {
		Parse(text)
	}
}

func BenchmarkMatch(b *testing.B) {
	m := Parse("*.log\nnode_modules/\nsrc/**/*.test.go\n!important.log")
	for b.Loop() {
		m.Match("src/pkg/handler_test.go", false)
	}
}

func BenchmarkMatchDeep(b *testing.B) {
	m := Parse("**/test/**/*.snap")
	for b.Loop() {
		m.Match("packages/core/src/test/fixtures/output.snap", false)
	}
}
