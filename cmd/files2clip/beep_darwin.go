package main

import "os/exec"

func beep() {
	exec.Command("osascript", "-e", "beep").Run()
}
