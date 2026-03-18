//go:build !darwin && !windows

package main

func beep() {
	// No audible beep on Linux/other — terminal bell is unreliable
}
