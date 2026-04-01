package clipboard

import (
	"os/exec"
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
		checkDarwinCmd(t, label, cmd, args)
	case "linux":
		checkLinuxCmd(t, label, cmd, args)
	case "windows":
		checkWindowsCmd(t, label, cmd, args)
	}
}

func checkDarwinCmd(t *testing.T, label, cmd string, args []string) {
	t.Helper()
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
}

func checkLinuxCmd(t *testing.T, label, cmd string, args []string) {
	t.Helper()
	switch cmd {
	case "xclip":
		checkXclipArgs(t, label, args)
	case "wl-copy":
		if label != "write" {
			t.Errorf("wl-copy should only be used for write, not %s", label)
		}
	case "wl-paste":
		checkWlPasteArgs(t, label, args)
	default:
		t.Errorf("%s: unexpected command %q on linux", label, cmd)
	}
}

func checkXclipArgs(t *testing.T, label string, args []string) {
	t.Helper()
	if len(args) < 2 {
		t.Fatalf("%s: expected at least 2 args, got %v", label, args)
	}
	if args[0] != "-selection" || args[1] != "clipboard" {
		t.Errorf("%s: expected -selection clipboard, got %v", label, args[:2])
	}
	if label == "read" && (len(args) != 3 || args[2] != "-o") {
		t.Errorf("%s: expected -o flag for read, got %v", label, args)
	}
}

func checkWlPasteArgs(t *testing.T, label string, args []string) {
	t.Helper()
	if label != "read" {
		t.Errorf("wl-paste should only be used for read, not %s", label)
	}
	if len(args) != 1 || args[0] != "--no-newline" {
		t.Errorf("%s: expected [--no-newline], got %v", label, args)
	}
}

func checkWindowsCmd(t *testing.T, label, cmd string, args []string) {
	t.Helper()
	if cmd != "powershell" {
		t.Errorf("%s: cmd = %q, want %q", label, cmd, "powershell")
	}
	if len(args) == 0 {
		t.Errorf("%s: expected args for powershell, got none", label)
	}
}

func TestLinuxCmd(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-only test")
	}

	t.Run("xclip fallback without WAYLAND_DISPLAY", func(t *testing.T) {
		t.Setenv("WAYLAND_DISPLAY", "")
		checkLinuxXclipFallback(t)
	})

	t.Run("wayland preferred when available", func(t *testing.T) {
		t.Setenv("WAYLAND_DISPLAY", "wayland-0")
		checkLinuxWaylandPreferred(t)
	})
}

func checkLinuxXclipFallback(t *testing.T) {
	t.Helper()
	cmd, args, err := linuxCmd("write")
	if err != nil {
		t.Fatal(err)
	}
	if cmd != "xclip" {
		t.Errorf("write: cmd = %q, want xclip", cmd)
	}
	if len(args) < 2 || args[0] != "-selection" {
		t.Errorf("write: unexpected args %v", args)
	}

	cmd, args, err = linuxCmd("read")
	if err != nil {
		t.Fatal(err)
	}
	if cmd != "xclip" {
		t.Errorf("read: cmd = %q, want xclip", cmd)
	}
	if len(args) < 3 || args[2] != "-o" {
		t.Errorf("read: unexpected args %v", args)
	}
}

func checkLinuxWaylandPreferred(t *testing.T) {
	t.Helper()
	checkWaylandMode(t, "write", "wl-copy")
	checkWaylandMode(t, "read", "wl-paste")
}

func checkWaylandMode(t *testing.T, mode, waylandCmd string) {
	t.Helper()
	cmd, _, err := linuxCmd(mode)
	if err != nil {
		t.Fatal(err)
	}
	lookCmd := waylandCmd
	if _, lookErr := exec.LookPath(lookCmd); lookErr == nil {
		if cmd != waylandCmd {
			t.Errorf("%s: cmd = %q, want %s (WAYLAND_DISPLAY set)", mode, cmd, waylandCmd)
		}
	} else {
		if cmd != "xclip" {
			t.Errorf("%s: cmd = %q, want xclip (%s not found)", mode, cmd, waylandCmd)
		}
	}
}

func TestLinuxCmdFallback(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux-only test")
	}

	// Even with WAYLAND_DISPLAY set, if wl-copy/wl-paste aren't in PATH,
	// should fall back to xclip
	t.Setenv("WAYLAND_DISPLAY", "wayland-0")
	t.Setenv("PATH", "/nonexistent")

	cmd, _, err := linuxCmd("write")
	if err != nil {
		t.Fatal(err)
	}
	if cmd != "xclip" {
		t.Errorf("write: cmd = %q, want xclip (fallback)", cmd)
	}

	cmd, _, err = linuxCmd("read")
	if err != nil {
		t.Fatal(err)
	}
	if cmd != "xclip" {
		t.Errorf("read: cmd = %q, want xclip (fallback)", cmd)
	}
}
