package style

import (
	"strings"
	"testing"
)

func TestFormatFunctions(t *testing.T) {
	// Save and restore colorEnabled to test both paths
	orig := colorEnabled
	defer func() { colorEnabled = orig }()

	tests := []struct {
		name string
		fn   func(string) string
		icon string
	}{
		{"OK", OK, IconOK},
		{"Skip", Skip, IconSkip},
		{"Fail", Fail, IconFail},
		{"Limit", Limit, IconLimit},
		{"Info", Info, IconInfo},
		{"Done", Done, IconDone},
	}

	for _, tt := range tests {
		t.Run(tt.name+"/no_color", func(t *testing.T) {
			colorEnabled = false
			checkFormatNoColor(t, tt.fn, tt.icon)
		})
		t.Run(tt.name+"/color", func(t *testing.T) {
			colorEnabled = true
			checkFormatColor(t, tt.fn, tt.icon)
		})
	}
}

func checkFormatNoColor(t *testing.T, fn func(string) string, icon string) {
	t.Helper()
	result := fn("test message")
	if !strings.Contains(result, icon) {
		t.Errorf("expected icon %q in %q", icon, result)
	}
	if !strings.Contains(result, "test message") {
		t.Errorf("expected message in %q", result)
	}
	if strings.Contains(result, "\033[") {
		t.Errorf("no ANSI codes expected when color disabled: %q", result)
	}
}

func checkFormatColor(t *testing.T, fn func(string) string, icon string) {
	t.Helper()
	result := fn("test message")
	if !strings.Contains(result, icon) {
		t.Errorf("expected icon %q in %q", icon, result)
	}
	if !strings.Contains(result, "test message") {
		t.Errorf("expected message in %q", result)
	}
	if !strings.Contains(result, "\033[") {
		t.Errorf("expected ANSI codes when color enabled: %q", result)
	}
	if !strings.HasSuffix(result, reset) {
		t.Errorf("expected reset suffix: %q", result)
	}
}

func TestDetectColor(t *testing.T) {
	t.Run("NO_COLOR set", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		if detectColor() {
			t.Error("detectColor should return false when NO_COLOR is set")
		}
	})

	t.Run("TERM dumb", func(t *testing.T) {
		t.Setenv("NO_COLOR", "")
		t.Setenv("TERM", "dumb")
		if detectColor() {
			t.Error("detectColor should return false when TERM=dumb")
		}
	})
}

func BenchmarkOK(b *testing.B) {
	colorEnabled = false
	for b.Loop() {
		OK("some/path/to/file.go")
	}
}

func BenchmarkOKColor(b *testing.B) {
	colorEnabled = true
	for b.Loop() {
		OK("some/path/to/file.go")
	}
}
