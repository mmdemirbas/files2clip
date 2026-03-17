package clipboard

import (
	"runtime"
	"testing"
)

func TestWriteCmd(t *testing.T) {
	cmd, args, err := writeCmd()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	switch runtime.GOOS {
	case "darwin":
		if cmd != "pbcopy" {
			t.Errorf("expected pbcopy, got %s", cmd)
		}
		if len(args) != 0 {
			t.Errorf("expected no args, got %v", args)
		}
	case "linux":
		if cmd != "xclip" {
			t.Errorf("expected xclip, got %s", cmd)
		}
	case "windows":
		if cmd != "powershell" {
			t.Errorf("expected powershell, got %s", cmd)
		}
	}
}

func TestReadCmd(t *testing.T) {
	cmd, args, err := readCmd()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	switch runtime.GOOS {
	case "darwin":
		if cmd != "pbpaste" {
			t.Errorf("expected pbpaste, got %s", cmd)
		}
		if len(args) != 0 {
			t.Errorf("expected no args, got %v", args)
		}
	case "linux":
		if cmd != "xclip" {
			t.Errorf("expected xclip, got %s", cmd)
		}
	case "windows":
		if cmd != "powershell" {
			t.Errorf("expected powershell, got %s", cmd)
		}
	}
}
