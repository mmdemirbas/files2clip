package clipboard

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
)

// Set writes data to the system clipboard.
func Set(data []byte) error {
	cmdName, args, err := writeCmd()
	if err != nil {
		return err
	}
	cmd := exec.Command(cmdName, args...)
	cmd.Stdin = bytes.NewReader(data)
	return cmd.Run()
}

// Get reads text data from the system clipboard.
func Get() ([]byte, error) {
	cmdName, args, err := readCmd()
	if err != nil {
		return nil, err
	}
	return exec.Command(cmdName, args...).Output()
}

func readCmd() (string, []string, error) {
	switch runtime.GOOS {
	case "windows":
		return "powershell", []string{
			"-NoProfile", "-NonInteractive",
			"-Command", "Get-Clipboard",
		}, nil
	case "darwin":
		return "pbpaste", []string{}, nil
	case "linux":
		return "xclip", []string{"-selection", "clipboard", "-o"}, nil
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
		return "pbcopy", []string{}, nil
	case "linux":
		return "xclip", []string{"-selection", "clipboard"}, nil
	default:
		return "", nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
