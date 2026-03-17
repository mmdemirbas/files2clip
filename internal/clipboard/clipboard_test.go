package clipboard

import (
	"runtime"
	"testing"
)

func TestWriteCmd(t *testing.T) {
	testClipboardCmd(t, "write", writeCmd)
}

func TestReadCmd(t *testing.T) {
	testClipboardCmd(t, "read", readCmd)
}

func testClipboardCmd(t *testing.T, label string, fn func() (string, []string, error)) {
	t.Helper()
	cmd, args, err := fn()
	if err != nil {
		t.Fatalf("%s: unexpected error: %v", label, err)
	}

	switch runtime.GOOS {
	case "darwin":
		wantCmd := "pbcopy"
		if label == "read" {
			wantCmd = "pbpaste"
		}
		if cmd != wantCmd {
			t.Errorf("%s: cmd = %q, want %q", label, cmd, wantCmd)
		}
		if len(args) != 0 {
			t.Errorf("%s: expected no args, got %v", label, args)
		}
	case "linux":
		if cmd != "xclip" {
			t.Errorf("%s: cmd = %q, want %q", label, cmd, "xclip")
		}
		if len(args) < 2 {
			t.Fatalf("%s: expected at least 2 args, got %v", label, args)
		}
		if args[0] != "-selection" || args[1] != "clipboard" {
			t.Errorf("%s: expected -selection clipboard, got %v", label, args[:2])
		}
		if label == "read" {
			if len(args) != 3 || args[2] != "-o" {
				t.Errorf("%s: expected -o flag for read, got %v", label, args)
			}
		}
	case "windows":
		if cmd != "powershell" {
			t.Errorf("%s: cmd = %q, want %q", label, cmd, "powershell")
		}
		if len(args) == 0 {
			t.Errorf("%s: expected args for powershell, got none", label)
		}
	}
}
