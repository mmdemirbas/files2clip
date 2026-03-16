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
