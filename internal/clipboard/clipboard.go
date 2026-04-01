package clipboard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// Set writes data to the system clipboard.
func Set(data []byte) error {
	cmdName, args, err := writeCmd()
	if err != nil {
		return err
	}
	cmd := exec.Command(cmdName, args...) //nolint:gosec // G204: cmdName is an OS-selected constant, not user input
	cmd.Stdin = bytes.NewReader(data)
	return cmd.Run()
}

// Get reads text data from the system clipboard.
func Get() ([]byte, error) {
	cmdName, args, err := readCmd()
	if err != nil {
		return nil, err
	}
	return exec.Command(cmdName, args...).Output() //nolint:gosec // G204: cmdName is an OS-selected constant, not user input
}

func readCmd() (string, []string, error) {
	switch runtime.GOOS {
	case "windows":
		return "powershell", []string{
			"-NoProfile", "-NonInteractive",
			"-Command", "Get-Clipboard",
		}, nil
	case "darwin":
		return "pbpaste", nil, nil
	case "linux":
		return linuxCmd("read")
	default:
		return "", nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func writeCmd() (string, []string, error) {
	switch runtime.GOOS {
	case "windows":
		return "powershell", []string{
			"-NoProfile", "-NonInteractive",
			"-Command", "[Console]::InputEncoding=[System.Text.Encoding]::UTF8; [Console]::OutputEncoding=[System.Text.Encoding]::UTF8; Set-Clipboard ([Console]::In.ReadToEnd())",
		}, nil
	case "darwin":
		return "pbcopy", nil, nil
	case "linux":
		return linuxCmd("write")
	default:
		return "", nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// linuxCmd returns the appropriate clipboard command for Linux,
// preferring Wayland (wl-copy/wl-paste) when available, falling
// back to X11 (xclip).
func linuxCmd(mode string) (string, []string, error) {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		if mode == "read" {
			if _, err := exec.LookPath("wl-paste"); err == nil {
				return "wl-paste", []string{"--no-newline"}, nil
			}
		} else {
			if _, err := exec.LookPath("wl-copy"); err == nil {
				return "wl-copy", nil, nil
			}
		}
	}

	// Fall back to xclip (X11 / XWayland)
	if mode == "read" {
		return "xclip", []string{"-selection", "clipboard", "-o"}, nil
	}
	return "xclip", []string{"-selection", "clipboard"}, nil
}
