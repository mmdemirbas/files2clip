package main

import "os/exec"

func beep() {
	_ = exec.Command("osascript", "-e", "beep").Run() //nolint:errcheck // best-effort audible notification
}
